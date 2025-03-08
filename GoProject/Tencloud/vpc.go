package Tencloud

import (
	"Tencloud/auth"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

var templateID = ""

func AddWhiteIP(ip, remark string) (string, error) {
	ipadd := matchIP(ip)
	if ipadd == "" {
		return "", errors.New("添加失败, 请检查IP地址格式是否正确")
	}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(auth.Kdl1, "ap-beijing", cpf)
	request := vpc.NewAddTemplateMemberRequest()
	request.TemplateId = common.StringPtr(templateID)
	request.TemplateMember = []*vpc.MemberInfo{
		{
			Member:      common.StringPtr(ipadd),
			Description: common.StringPtr(remark),
		},
	}
	_, err := client.AddTemplateMember(request)
	if err != nil {
		return "", err
	}
	return "添加白名单IP: " + ip + " 成功.", nil
}

func DeleteWhiteIP(ip string) (string, error) {
	ipadd := matchIP(ip)
	if ipadd == "" {
		return "", errors.New("删除失败, 请检查IP地址格式是否正确")
	}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(auth.Kdl1, "ap-beijing", cpf)
	request := vpc.NewDeleteTemplateMemberRequest()

	request.TemplateId = common.StringPtr("ipm-aoaf7uj8")
	request.TemplateMember = []*vpc.MemberInfo{
		{
			Member: common.StringPtr(ipadd),
		},
	}
	_, err := client.DeleteTemplateMember(request)
	if err != nil {
		return "", err
	}
	return "删除白名单IP: " + ip + " 成功.", nil
}

func GetWhiteIP() (string, error) {
	ipData := make([]string, 0, 20)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(auth.Kdl1, "ap-beijing", cpf)
	request := vpc.NewDescribeAddressTemplatesRequest()

	request.Filters = []*vpc.Filter{
		{
			Name:   common.StringPtr("address-template-id"),
			Values: common.StringPtrs([]string{templateID}),
		},
	}
	request.Limit = common.StringPtr("100")
	response, err := client.DescribeAddressTemplates(request)
	if err != nil {
		return "", err
	}

	for _, i := range response.Response.AddressTemplateSet[0].AddressExtraSet {
		ipData = append(ipData, fmt.Sprintf("%s 备注: %s", *i.Address, *i.Description))
	}
	return strings.Join(ipData, "\n"), nil
}

func matchIP(ipStr string) (ip string) {
	re := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:/\d{1,2})?`)
	ip = re.FindString(ipStr)
	return
}
