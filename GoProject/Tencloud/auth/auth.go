package auth

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type TenCloudKey struct {
	AccessID  string
	AccessKey string
}

var (
	tenCloud1 TenCloudKey = TenCloudKey{"", ""}
	tenCloud2 TenCloudKey = TenCloudKey{"", ""}
	cwtData   TenCloudKey = TenCloudKey{"", ""}
	Kdl1                  = common.NewCredential(tenCloud1.AccessID, tenCloud1.AccessKey)
	Kdl2                  = common.NewCredential(tenCloud2.AccessID, tenCloud2.AccessKey)
	Cwt                   = common.NewCredential(cwtData.AccessID, cwtData.AccessKey)
)
