package main

import (
	"kubernetesDev/api"
	"opsServer/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(auth.AUTH())

	r.POST("/update", api.Update)

	r.Run(":8072")
}
