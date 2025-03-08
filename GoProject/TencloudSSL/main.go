package main

import (
	"fmt"
	"time"

	"Aliyun"
	"util"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

var (
	logFile   = "/data/kdl/log/devops/checkSSLCertificate.log"
	logInfo   = util.LogConf(logFile, "[INFO] ")
	logError  = util.LogConf(logFile, "[ERROR] ")
	accessId  = ""
	accessKey = ""
	client    *ssl.Client
	emaile    = ""
	title     = "[SSL证书通知]"
	notifyUrl = ""
)

func init() {
	credential := common.NewCredential(
		accessId,
		accessKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	client, _ = ssl.NewClient(credential, "", cpf)
}

func main() {
	logInfo.Println("run...")
	d, e := getData()
	if e != nil {
		logError.Println(e)
		return
	}
	for _, x := range d.Response.Certificates {
		if *x.IsDv && x.CertEndTime != nil {
			r := isCreate(*x.CertEndTime)
			if r {
				err := createSSLCertificate(*x.Domain)
				if err != nil {
					logError.Println(err)
				}
			}
		}
	}
}

func getData() (*ssl.DescribeCertificatesResponse, error) {
	request := ssl.NewDescribeCertificatesRequest()
	request.CertificateType = common.StringPtr("SVR")
	request.Limit = common.Uint64Ptr(100)
	response, err := client.DescribeCertificates(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, err
	}
	return response, nil
}

func isCreate(t string) (r bool) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timeObj, err := time.ParseInLocation(time.DateTime, t, loc)
	if err != nil {
		return false
	}
	now := time.Now()
	tmpTime := now.Sub(timeObj)
	days := int(tmpTime.Hours() / 24)
	if days == -2 {
		r = true
	}
	return
}

func createSSLCertificate(domain string) error {
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr("DNS")
	request.DomainName = common.StringPtr(domain)
	request.PackageType = common.StringPtr("83")
	request.ContactEmail = common.StringPtr(emaile)
	request.ValidityPeriod = common.StringPtr("RSA")
	request.Alias = common.StringPtr(domain + "_new")

	response, err := client.ApplyCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return err
	}
	authSSLCertificateInfo(*response.Response.CertificateId)
	url, err := getSSLCertificateStatus(*response.Response.CertificateId)
	if err != nil {
		util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s证书生成失败！", domain)})
	} else {
		util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s证书生成成功 --->  %s", domain, url)})
	}
	return nil
}

func authSSLCertificateInfo(ID string) {
	request := ssl.NewDescribeCertificateRequest()
	request.CertificateId = common.StringPtr(ID)
	response, err := client.DescribeCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logError.Printf("An API error has returned: %s", err)
		return
	}
	data := response.Response.DvAuthDetail.DvAuths[0]
	err = Aliyun.AddDomainRecord(*data.DvAuthSubDomain, *data.DvAuthDomain, "CNAME", *data.DvAuthValue, "default", 600)
	if err != nil {
		logError.Println(err)
		return
	}
}

func getUrl(ID string) (string, error) {
	request := ssl.NewDescribeDownloadCertificateUrlRequest()

	request.CertificateId = common.StringPtr(ID)
	request.ServiceType = common.StringPtr("nginx")

	response, err := client.DescribeDownloadCertificateUrl(request)
	if err != nil {
		return "", err
	}

	return *response.Response.DownloadCertificateUrl, nil
}

func getSSLCertificateStatus(ID string) (string, error) {
	for range 15 {
		request := ssl.NewDescribeCertificateDetailRequest()

		request.CertificateId = common.StringPtr(ID)

		response, err := client.DescribeCertificateDetail(request)
		if _, ok := err.(*errors.TencentCloudSDKError); ok {
			return "", err
		}
		var url string
		activeAuth(ID)
		if *response.Response.Status == 1 {
			url, err = getUrl(ID)
			if err != nil {
				return "", err
			}
			return url, nil
		} else {
			time.Sleep(1 * time.Minute)
		}
	}
	return "", fmt.Errorf("%s ssl证书创建超时", ID)
}

func activeAuth(ID string) {
	request := ssl.NewCompleteCertificateRequest()
	request.CertificateId = common.StringPtr(ID)
	_, err := client.CompleteCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		logError.Println(err)
	}
}
