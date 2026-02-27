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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup payload", "code": "INVALID_PAYLOAD"})
		return
	}

	body.NewStudentRole = "student"

	if body.NewStudentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}
	if err := services.CreateStudent(body, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Successfully Created", "code": "ACCOUNT_CREATED"})
}

func CreateStaff(c *gin.Context) {
	var body models.NewStaff
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup payload", "code": "INVALID_PAYLOAD"})
		return
	}

	if body.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}
	if err := services.CreateStaff(body, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Successfully Created", "code": "ACCOUNT_CREATED"})
}

func Login(c *gin.Context) {
	var body models.UserLogin
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload", "code": "INVALID_LOGIN_PAYLOAD"})
		return
	}
	resp, err := services.LoginUser(body, c)

	if err != nil {
		if err.Error() == "PASSWORD_REQUIRED" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "PASSWORD_REQUIRED",
				"message": "Password is required",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid ID or password",
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func UpdateStudent(c *gin.Context) {
	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required", "code": "ID_REQUIRED"})
		return
	}

	var body models.UpdateStudent
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload", "code": "INVALID_PAYLOAD"})
		return
	}

	if err := services.UpdateStudent(studentID, body, c); err != nil {
		if err.Error() == "NO_FIELDS_TO_UPDATE" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update", "code": "INVALID_PAYLOAD"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update student", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Student updated successfully", "code": "OPERATION_SUCCESSFUL"})
}

func DeleteStudent(c *gin.Context) {
	studentID := c.Param("studentID")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student ID is required", "code": "ID_REQUIRED"})
		return
	}

	if err := services.DeleteStudent(studentID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete student", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully", "code": "OPERATION_SUCCESSFUL"})
}

func UpdateStaff(c *gin.Context) {
	staffID := c.Param("staffID")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "staff ID is required", "code": "ID_REQUIRED"})
		return
	}

	var body models.UpdateStaff
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload", "code": "INVALID_PAYLOAD"})
		return
	}

	if err := services.UpdateStaff(staffID, body, c); err != nil {
		if err.Error() == "NO_FIELDS_TO_UPDATE" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update", "code": "INVALID_PAYLOAD"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update staff", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Staff updated successfully", "code": "OPERATION_SUCCESSFUL"})
}

func DeleteStaff(c *gin.Context) {
	staffID := c.Param("staffID")
	if staffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "staff ID is required", "code": "ID_REQUIRED"})
		return
	}

	if err := services.DeleteStaff(staffID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete staff", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Staff deleted successfully", "code": "OPERATION_SUCCESSFUL"})
}

func GetStudentByID(c *gin.Context) {
	userID := c.Param("userId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required", "code": "ID_REQUIRED"})
		return
	}

	user, err := services.GetStudentByID(userID, c)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found", "code": "NOT_FOUND"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get student", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetStaffByID(c *gin.Context) {
	userID := c.Param("userId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required", "code": "ID_REQUIRED"})
		return
	}

	user, err := services.GetStaffByID(userID, c)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found", "code": "NOT_FOUND"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get staff", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetStaff(c *gin.Context) {
	students, err := services.GetAllStaff(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get staff", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, students)

}
