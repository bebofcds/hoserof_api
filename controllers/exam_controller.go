package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"HOSEROF_API/models"
	"HOSEROF_API/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	token := c.MustGet("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	adminID := claims["user_ID"].(string)

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
	id, err := services.CreateExam(exam, body.Questions)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create exam"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "exam_id": id})
}

func ListExamsForStudent(c *gin.Context) {
	token := c.MustGet("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userClass, ok := claims["user_class"].(string)
	studentID := claims["user_ID"].(string)
	if !ok || userClass == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_class missing from token",
		})
		return
	}

	class := userClass
	exams, err := services.GetExamsForClass(class, studentID)
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
	qs, err := services.GetExamQuestions(examID, true)
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
	token := c.MustGet("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	studentID := claims["user_ID"].(string)

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

	err := services.SubmitExam(examID, studentID, parsed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetSubmissionsForExam(c *gin.Context) {
	examID := c.Param("examID")
	subs, err := services.GetAllSubmissions(examID)
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

func GradeAnswer(c *gin.Context) {
	var body GradeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	examID := c.Param("examID")
	if err := services.GradeWrittenAnswer(examID, body.StudentID, body.QID, body.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ReleaseResultsHandler(c *gin.Context) {
	examID := c.Param("examID")
	if err := services.ReleaseResults(examID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release results"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetReleasedResultForStudent(c *gin.Context) {
	token := c.MustGet("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	studentID := claims["user_ID"].(string)

	examID := c.Param("examID")

	result, err := services.GetReleasedResult(examID, studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
func ListReleasedResults(c *gin.Context) {
	token := c.MustGet("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	studentID := claims["user_ID"].(string)

	results, err := services.GetAllReleasedResultsForStudent(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load results"})
		return
	}
	if results == nil {
		results = []models.ResultSummary{}
	}
	c.JSON(http.StatusOK, results)
}
