package controllers

import (
	"net/http"

	"HOSEROF_API/models"
	"HOSEROF_API/services"

	"github.com/gin-gonic/gin"
)

func CreateStudent(c *gin.Context) {
	var body models.NewUser
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup payload", "details": err.Error()})
		return
	}

	body.NewStudentRole = "student"

	if body.NewStudentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}
	if err := services.CreateStudent(body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Successfully Created", "code": "ACCOUNT_CREATED"})
}

func CreateStaff(c *gin.Context) {
	var body models.NewStaff
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup payload", "details": err.Error()})
		return
	}

	if body.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}
	if err := services.CreateStaff(body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Successfully Created", "code": "ACCOUNT_CREATED"})
}

func Login(c *gin.Context) {
	var body models.UserLogin
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}
	resp, err := services.LoginUser(body)

	if err != nil {
		if err.Error() == "PASSWORD_REQUIRED" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "PASSWORD_REQUIRED",
				"message": "Password is required",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_LOGIN",
			"message": "Invalid ID or password",
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func TokenCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": "token valid"})
}
