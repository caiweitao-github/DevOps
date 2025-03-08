package api

import (
	"kubernetesDev/kubeletutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Update(c *gin.Context) {
	var kdl kubeletutil.KdlConfig
	if err := c.ShouldBindJSON(&kdl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}
	err := kdl.UpdateDeploy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Restart(c *gin.Context) {
	var kdl kubeletutil.KdlConfig
	if err := c.ShouldBindJSON(&kdl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data error",
		})
		return
	}
	err := kdl.RestartDeployment()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "update fail.",
		})
	}
}
