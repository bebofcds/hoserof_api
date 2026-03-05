package controllers

import (
	"HOSEROF_API/middleware"
	"HOSEROF_API/models"
	"HOSEROF_API/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func MarkAttendance(c *gin.Context) {
	var body struct {
		StudentID string `json:"studentId"`
		Attended  bool   `json:"attended"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}
	if body.StudentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "studentId required", "code": "ID_REQUIRED"})
		return
	}
	err := services.MarkAttendance(body.StudentID, body.Attended, c)

	if err != nil {

		if err.Error() == "no user found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no user found", "code": "NOT_FOUND"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save attendance", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})

}

// func MarkAttendanceManual(c *gin.Context) {

// 	var body struct {
// 		StudentID string `json:"studentId"`
// 		Date      string `json:"date"`
// 		Attended  bool   `json:"attended"`
// 	}

// 	if err := c.ShouldBindJSON(&body); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
// 		return
// 	}

// 	if body.StudentID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "student id required"})
// 		return
// 	}

// 	if body.Date == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "date required (YYYY-MM-DD)"})
// 		return
// 	}

// 	err := services.MarkAttendanceManual(body.StudentID, body.Date, body.Attended)
// 	if err != nil {
// 		if err.Error() == "no user found" {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "no user found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"success": true})
// }

func GetAttendance(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	resp, err := services.GetAttendance(studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get attendance", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetAttendanceByID(c *gin.Context) {
	studentID := c.Param("studentID")

	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}
	resp, err := services.GetAttendance(studentID, c)

	if err != nil {

		if err.Error() == "no user found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no user found", "code": "NOT_FOUND"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get attendance", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, resp)

}

func GetStudentsByClass(c *gin.Context) {
	classID := c.Param("classId")
	hideMarkedToday := c.Query("hideMarkedToday") == "true"

	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "classId is required", "code": "ID_REQUIRED"})
		return
	}

	students, err := services.GetStudents(classID, hideMarkedToday, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get students", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, students)

}

func GetStudentsForTeacher(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	classID := claims.UserClass

	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "classId is required", "code": "ID_REQUIRED"})
		return
	}

	students, err := services.GetStudents(classID, false, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get students", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, students)

}

func MarkAttendanceBatch(c *gin.Context) {
	var body []models.AttendanceBatchRecord

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload", "code": "INVALID_PAYLOAD"})
		return
	}

	err := services.MarkAttendanceBatch(body, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save attendance", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetClassAttendanceSummary(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	classID := claims.UserClass
	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "classId is required",
			"code":  "ID_REQUIRED",
		})
		return
	}

	resp, err := services.GetClassAttendanceSummary(classID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get attendance summary",
			"code":  "SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
