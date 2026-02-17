package controllers

import (
	"fmt"
	"log"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
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
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create exam"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "exam_id": id})
}

func ListExamsForStudent(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	userClass := claims.UserClass
	studentID := claims.ID
	if userClass == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_class missing from token",
		})
		return
	}

	class := userClass
	exams, err := services.GetExamsForClass(class, studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exams == nil {
		exams = []models.Exam{}
	}

	c.JSON(http.StatusOK, exams)
}

func GetExamForStudent(c *gin.Context) {
	examID := c.Param("examID")
	qs, err := services.GetExamQuestions(examID, true, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get exam"})
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

	examID := c.Param("examID")
	var body SubmitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetSubmissionsForExam(c *gin.Context) {
	examID := c.Param("examID")
	subs, err := services.GetAllSubmissions(examID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submissions"})
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
	examID := c.Param("examID")

	if err := services.DeleteExam(examID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ReleaseResultsHandler(c *gin.Context) {
	examID := c.Param("examID")
	if err := services.ReleaseResults(examID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release results"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetReleasedResultForStudent(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	examID := c.Param("examID")

	result, err := services.GetReleasedResult(examID, studentID, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
func ListReleasedResults(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	studentID := claims.ID

	results, err := services.GetAllReleasedResultsForStudent(studentID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load results"})
		return
	}
	if results == nil {
		results = []models.ResultSummary{}
	}
	c.JSON(http.StatusOK, results)
}
