package api

import (
	"opsServer/rpcClient"
	"opsServer/service"

	"github.com/gin-gonic/gin"
)

var client *rpcClient.RpcClient

func init() {
	var err error
	client, err = rpcClient.DialService(service.RpcConfig.Protocol, service.RpcConfig.Address)
	if err != nil {
		panic(err)
	}
}

func Gettransfer(c *gin.Context) {
	var res []string
	err := client.GetNode(&res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Getline(c *gin.Context) {
	var pa service.Parameter
	if err := c.ShouldBindJSON(&pa); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	var res []service.Result
	err := client.GetData(pa, &res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Report(c *gin.Context) {
	var p service.NodeList
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.CheckStatus(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func Getorder(c *gin.Context) {
	var res []service.OrderDtat
	err := client.GetOrderData(&res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Getmonitorsfps(c *gin.Context) {
	var res []service.SfpsData
	err := client.GetSfpsData(&res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Sfpsmonitorreport(c *gin.Context) {
	var p service.NodeList
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.CheckSfpsStatus(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func Queryorderid(c *gin.Context) {
	p := struct {
		SecretID string `json:"secret_id"`
	}{}
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	var res string
	err := client.QueryOrderID(p.SecretID, &res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Unforbidip(c *gin.Context) {
	var p service.ReportIP
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	var res []string
	err := client.UnForbidIP(p, &res)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, res)
}

func Reportforbidip(c *gin.Context) {
	var p service.NginxForbidData
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.ReportForbidIP(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func Updatedomainrecord(c *gin.Context) {
	var p service.DomainRecordDate
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.UpdataTpsDomainRecord(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func Reportjipfdata(c *gin.Context) {
	var p service.JipFDate
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.UpdateCache(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func Modifytdpsstatus(c *gin.Context) {
	var p service.DataEntry
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.ModifyTdpsStatus(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func ReportIpPool(c *gin.Context) {
	var p service.IpPool
	if err := c.ShouldBindJSON(&p); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	err := client.ReportIpPool(p)
	if err != nil {
		service.Fail[string](c, service.ServerError)
		return
	}
	service.OK(c, "")
}

func GetCabNode(c *gin.Context) {
	var res []string
	err := client.GetCabNode(&res)
	if err != nil {
		service.FailWithMsg[string](c, service.ServerError, err.Error())
		return
	}
	service.OK(c, res)
}

func GetCabLine(c *gin.Context) {
	var pa service.Parameter
	if err := c.ShouldBindJSON(&pa); err != nil {
		service.Fail[string](c, service.InvalidArgs)
		return
	}
	var res []service.Result
	err := client.GetCabLine(pa, &res)
	if err != nil {
		service.FailWithMsg[string](c, service.ServerError, err.Error())
		return
	}
	service.OK(c, res)
}
