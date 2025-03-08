package Tencloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var (
	client    *cvm.Client
	cpf       *profile.ClientProfile
	accessId  = ""
	accessKey = ""
	regionId  = "ap-beijing"
	endPoint  = "cvm.tencentcloudapi.com"
)

var (
	nodeType               = "POSTPAID_BY_HOUR"
	nodeZone               = "ap-beijing-6"
	nodeConf               = "S6.2XLARGE16"
	nodeVpc                = "vpc-4xhq1hye"
	nodeSubnetId           = "subnet-qi8yh2uv"
	bandwidthType          = "BANDWIDTH_POSTPAID_BY_HOUR"
	maxBandwidthOut  int64 = 100
	securityGroupIds       = "sg-pg5zjchl"
)

func init() {
	var err error
	client, err = initClient()
	if err != nil {
		panic(err)
	}
}

func initClient() (*cvm.Client, error) {
	cpf = profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endPoint
	credential := common.NewCredential(accessId, accessKey)
	client, err := cvm.NewClient(credential, regionId, cpf)
	return client, err
}

func createCvm() (string, error) {
	nodeImg, err := getImgId()
	if err != nil {
		return "", err
	}
	request := cvm.NewRunInstancesRequest()
	request.InstanceChargeType = common.StringPtr(nodeType)
	request.Placement = &cvm.Placement{
		Zone:      common.StringPtr(nodeZone),
		ProjectId: common.Int64Ptr(0),
	}
	request.InstanceType = common.StringPtr(nodeConf)
	request.ImageId = common.StringPtr(nodeImg)
	request.SystemDisk = &cvm.SystemDisk{
		DiskType: common.StringPtr("CLOUD_BSSD"),
		DiskSize: common.Int64Ptr(100),
	}
	request.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:            common.StringPtr(nodeVpc),
		SubnetId:         common.StringPtr(nodeSubnetId),
		AsVpcGateway:     common.BoolPtr(false),
		Ipv6AddressCount: common.Uint64Ptr(0),
	}
	request.InternetAccessible = &cvm.InternetAccessible{
		InternetChargeType:      common.StringPtr(bandwidthType),
		InternetMaxBandwidthOut: common.Int64Ptr(maxBandwidthOut),
		PublicIpAssigned:        common.BoolPtr(true),
	}
	request.InstanceCount = common.Int64Ptr(1)
	request.LoginSettings = &cvm.LoginSettings{
		KeepImageLogin: common.StringPtr("true"),
	}
	request.SecurityGroupIds = common.StringPtrs([]string{securityGroupIds})
	request.EnhancedService = &cvm.EnhancedService{
		SecurityService: &cvm.RunSecurityServiceEnabled{
			Enabled: common.BoolPtr(true),
		},
		MonitorService: &cvm.RunMonitorServiceEnabled{
			Enabled: common.BoolPtr(true),
		},
	}
	request.DisableApiTermination = common.BoolPtr(false)
	response, err := client.RunInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return "", err
	}
	return string(*response.Response.InstanceIdSet[0]), nil
}

func getNodeStatus(id string) (string, error) {
	request := cvm.NewDescribeInstancesStatusRequest()

	request.InstanceIds = common.StringPtrs([]string{id})
	response, err := client.DescribeInstancesStatus(request)
	if err != nil {
		return "", fmt.Errorf("get node status error: %v", err)
	}
	return string(*response.Response.InstanceStatusSet[0].InstanceState), nil
}

func GetNodeIp() (string, string, error) {
	id, err := createCvm()
	if err != nil {
		return "", "", fmt.Errorf("create cvm error: %v", err)
	}
	i := 1
	for i < 15 {
		status, err := getNodeStatus(id)
		if err != nil {
			return "", "", fmt.Errorf("get node status error: %v", err)
		}
		if status == "RUNNING" {
			break
		} else {
			i++
			time.Sleep(time.Second * 3)
			continue
		}
	}
	request := cvm.NewDescribeInstancesRequest()
	request.InstanceIds = common.StringPtrs([]string{id})
	response, err := client.DescribeInstances(request)
	if err != nil {
		return "", "", fmt.Errorf("get node ip error: %v", err)
	}
	return id, string(*response.Response.InstanceSet[0].PublicIpAddresses[0]), nil
}

func DeleteServer(id string) error {
	request := cvm.NewTerminateInstancesRequest()
	request.InstanceIds = common.StringPtrs([]string{id})
	_, err := client.TerminateInstances(request)
	if err != nil {
		return err
	}
	return nil
}

func getImgId() (string, error) {
	re, _ := regexp.Compile("tpsbak-.*")
	request := cvm.NewDescribeImagesRequest()
	request.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("image-type"),
			Values: common.StringPtrs([]string{"PRIVATE_IMAGE"}),
		},
	}
	response, err := client.DescribeImages(request)
	if err != nil {
		return "", err
	}
	for _, v := range response.Response.ImageSet {
		if re.MatchString(*v.ImageName) {
			return *v.ImageId, nil
		}
	}
	return "", fmt.Errorf("not found tps-bak image")
}
