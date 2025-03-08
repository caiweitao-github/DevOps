package main

import (
	"fmt"
	"time"
	"util"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

var (
	logFile                 = "/data/kdl/log/devops/nginxBandwidthCheck.log"
	logInfo                 = util.LogConf(logFile, "[INFO] ")
	logError                = util.LogConf(logFile, "[ERROR] ")
	accessId                = ""
	accessKey               = ""
	monitorIndicators       = ""
	credential              = common.NewCredential(accessId, accessKey)
	filtrKey                = "wwwnginx"
	maxThreshold      int64 = 50
	title                   = "[Nginx带宽升级通知]"
	notifyUrl               = ""
)

type nodeData struct {
	serverID  string
	bandwidth int64
}

func main() {
	logInfo.Println("run...")
	m, err := getNodeID()
	if err != nil {
		logError.Fatalf("get node err: %v", err)
	}
	for _, v := range m {
		if v.bandwidth >= maxThreshold {
			return
		}
		b, err := getBandwidthData(v.serverID)
		if err != nil {
			logError.Fatalf("get bandwidth err: %v", err)
		}
		if b <= float64(v.bandwidth)*0.8 {
			return
		}
	}
	for name, node := range m {
		var nodeBandwidth int64 = 0
		if node.bandwidth+5 >= maxThreshold {
			nodeBandwidth = maxThreshold
		} else {
			nodeBandwidth = nodeBandwidth + 5
		}
		err := updateBandwidth(node.serverID, nodeBandwidth)
		if err != nil {
			util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s带宽调整失败.", name)})
		} else {
			util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s带宽调整成功: %d ---> %d", name, node.bandwidth, nodeBandwidth)})
		}
	}
}

func getBandwidthData(serverID string) (float64, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"

	client, _ := monitor.NewClient(credential, "ap-beijing", cpf)

	request := monitor.NewGetMonitorDataRequest()

	request.Namespace = common.StringPtr("QCE/CVM")
	request.MetricName = common.StringPtr(monitorIndicators)
	request.Period = common.Uint64Ptr(300)
	request.Instances = []*monitor.Instance{
		{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("InstanceId"),
					Value: common.StringPtr(serverID),
				},
			},
		},
	}
	request.SpecifyStatistics = common.Int64Ptr(1)
	response, err := client.GetMonitorData(request)
	if err != nil {
		return 0., err
	}
	d := response.Response.DataPoints[0].AvgValues
	return *d[len(d)-1], nil
}

func getNodeID() (map[string]nodeData, error) {
	n := make(map[string]nodeData, 5)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	client, _ := cvm.NewClient(credential, "ap-beijing", cpf)
	request := cvm.NewDescribeInstancesRequest()

	request.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("tag-key"),
			Values: common.StringPtrs([]string{filtrKey}),
		},
	}
	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	d := response.Response.InstanceSet
	for _, i := range d {
		n[*i.InstanceName] = nodeData{*i.InstanceId, *i.InternetAccessible.InternetMaxBandwidthOut}
	}
	return n, nil
}

func updateBandwidth(serverID string, bandwidth int64) error {
	ti := time.Now().Format(time.DateOnly)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	client, _ := cvm.NewClient(credential, "ap-beijing", cpf)
	request := cvm.NewResetInstancesInternetMaxBandwidthRequest()

	request.InstanceIds = common.StringPtrs([]string{serverID})
	request.InternetAccessible = &cvm.InternetAccessible{
		InternetMaxBandwidthOut: common.Int64Ptr(bandwidth),
	}
	request.StartTime = common.StringPtr(ti)
	request.EndTime = common.StringPtr(ti)
	_, err := client.ResetInstancesInternetMaxBandwidth(request)
	if err != nil {
		return err
	}
	return nil
}
