package controllers

import (
	"HOSEROF_API/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTeacherDashboard(c *gin.Context) {
	resp, err := services.GetTeacherDashboard(c)
	if err != nil {
		log.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load dashboard",
			"code":  "SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetAdminDashboard(c *gin.Context) {
	resp, err := services.GetAdminDashboard(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load dashboard",
			"code":  "SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
