package main

import (
	"opsServer/api"
	"opsServer/auth"
	"opsServer/service"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/gettoken", auth.GetToken)

	r.Use(auth.AUTH())

	r.GET("/gettransfer", api.Gettransfer)

	r.POST("/getline", api.Getline)

	r.POST("/report", api.Report)

	r.GET("/getorder", api.Getorder)

	r.GET("/getmonitorsfps", api.Getmonitorsfps)

	r.POST("/sfpsmonitorreport", api.Sfpsmonitorreport)

	r.POST("/queryorderid", api.Queryorderid)

	r.POST("/unforbidip", api.Unforbidip)

	r.POST("/reportforbidip", api.Reportforbidip)

	r.POST("/updatedomainrecord", api.Updatedomainrecord)

	r.POST("/reportjipfdata", api.Reportjipfdata)

	r.POST("/modifytdpsstatus", api.Modifytdpsstatus)

	r.POST("/reportippool", api.ReportIpPool)

	r.GET("/getcabnode", api.GetCabNode)

	r.POST("/getcabline", api.GetCabLine)

	r.Run(service.GinConfig.Address)
}
