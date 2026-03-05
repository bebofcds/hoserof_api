package controllers

import (
	"fmt"
	"net/http"
	"time"

	"HOSEROF_API/middleware"
	"HOSEROF_API/models"
	"HOSEROF_API/services"

	"github.com/gin-gonic/gin"
)

type CreateExamBody struct {
	Title            string            `json:"title"`
	Class            string            `json:"class"`
	TimeLimitMinutes int               `json:"time_limit_minutes"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          time.Time         `json:"end_time"`
	Questions        []models.Question `json:"questions"`
}

func CreateExam(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	adminID := claims.ID

	var body CreateExamBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}

	exam := models.Exam{
		Title:            body.Title,
		Class:            body.Class,
		TimeLimitMinutes: body.TimeLimitMinutes,
		StartTime:        body.StartTime,
		EndTime:          body.EndTime,
		CreatedBy:        adminID,
	}
	id, err := services.CreateExam(exam, body.Questions, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create exam", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "exam_id": id})
}

func ListExamsForStudent(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	userClass := claims.UserClass
	studentID := claims.ID

	exams, err := services.GetExamsForClass(userClass, studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get exams for this student", "code": "SERVER_ERROR"})
		return
	}

	if exams == nil {
		exams = []models.Exam{}
	}

	c.JSON(http.StatusOK, exams)
}

func ListAllExams(c *gin.Context) {

	exams, err := services.GetAllExamsForAdmin(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get all exams", "code": "SERVER_ERROR"})
		return
	}

	if exams == nil {
		exams = []models.Exam{}
	}

	c.JSON(http.StatusOK, exams)
}

func GetExamForStudent(c *gin.Context) {
	examID := c.Param("exam_id")
	qs, err := services.GetExamQuestions(examID, true, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get exam", "code": "SERVER_ERROR"})
		return
	}

	if qs == nil {
		qs = []models.Question{}
	}

	c.JSON(http.StatusOK, qs)
}

func GetExamForAdmin(c *gin.Context) {
	examID := c.Param("exam_id")
	qs, err := services.GetExamQuestions(examID, false, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get exam", "code": "SERVER_ERROR"})
		return
	}

	if qs == nil {
		qs = []models.Question{}
	}

	c.JSON(http.StatusOK, qs)
}

type SubmitBody struct {
	Answers map[string]interface{} `json:"answers"`
}

func SubmitExam(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	examID := c.Param("exam_id")
	var body SubmitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}

	parsed := make(map[string]models.Answer)
	for qid, raw := range body.Answers {
		parsed[qid] = models.Answer{
			QID:      qid,
			Response: fmt.Sprintf("%v", raw),
		}
	}

	err := services.SubmitExam(examID, studentID, parsed, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit exam", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetSubmissionsForExam(c *gin.Context) {
	examID := c.Param("examID")
	subs, err := services.GetAllSubmissions(examID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submissions", "code": "SERVER_ERROR"})
		return
	}
	if subs == nil {
		subs = []models.Submission{}
	}
	c.JSON(http.StatusOK, subs)
}

type GradeRequest struct {
	StudentID string  `json:"student_id"`
	QID       string  `json:"qid"`
	Score     float64 `json:"score"`
}

func DeleteExam(c *gin.Context) {
	examID := c.Param("exam_id")

	if err := services.DeleteExam(examID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete exam", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ReleaseResults(c *gin.Context) {
	examID := c.Param("exam_id")
	if err := services.ReleaseResults(examID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release results", "code": "SERVER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetReleasedResultForStudent(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	examID := c.Param("exam_id")

	result, err := services.GetReleasedResult(examID, studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "faild to get released results", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func ListReleasedResults(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	results, err := services.GetAllReleasedResultsForStudent(studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load results", "code": "SERVER_ERROR"})
		return
	}
	if results == nil {
		results = []models.ResultSummary{}
	}
	c.JSON(http.StatusOK, results)
}

func GetStudentSubmittedExams(c *gin.Context) {
	studentID := c.Param("student_id")
	examID := c.Query("exam_id")

	// If exam_id is provided, return detailed results for that exam
	if examID != "" {
		result, err := services.GetStudentExamResultDetails(studentID, examID, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get submitted exams", "code": "SERVER_ERROR"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// Otherwise, return list of submitted exams
	exams, err := services.GetStudentSubmittedExams(studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submitted exams", "code": "SERVER_ERROR"})
		return
	}
	if exams == nil {
		exams = []models.Exam{}
	}
	c.JSON(http.StatusOK, exams)
}
