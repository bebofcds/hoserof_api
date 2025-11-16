package controllers

import (
	"net/http"

	"HOSEROF_API/models"
	"HOSEROF_API/services"

	"github.com/gin-gonic/gin"
)

func Signup(c *gin.Context) {
	var body models.NewUser

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup payload"})
		return
	}

	if body.NewStudentID == "" || body.NewStudentPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and password required"})
		return
	}
	if err := services.SignupUser(body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Account Successfully Created"})
}

func Login(c *gin.Context) {
	var body models.UserLogin
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}
	resp, err := services.LoginUser(body)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func TokenCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": "token valid"})
}
