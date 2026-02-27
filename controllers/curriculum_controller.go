package controllers

import (
	"HOSEROF_API/middleware"
	"HOSEROF_API/models"
	"HOSEROF_API/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadCurriculumBody struct {
	ClassID string `form:"class_id" binding:"required"`
	Title   string `form:"title" binding:"required"`
}

func UploadCurriculum(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.Claims)
	userID := claims.ID

	var body UploadCurriculumBody
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "code": "FILE_REQUIRED"})
		return
	}
	defer file.Close()

	if header.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size must be less than 50MB", "code": "FILE_LIMIT"})
		return
	}

	req := models.UploadCurriculumRequest{
		ClassID: body.ClassID,
		Title:   body.Title,
	}

	curriculum, err := services.UploadCurriculum(c.Request.Context(), req, file, header, userID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload curriculum", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"curriculum": curriculum,
	})
}

func GetCurriculumsByClass(c *gin.Context) {
	classID := c.Param("class_id")
	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required", "code": "ID_REQUIRED"})
		return
	}

	curriculums, err := services.GetCurriculumsByClass(c.Request.Context(), classID, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get curriculums", "code": "SERVER_ERROR"})
		return
	}

	if curriculums == nil {
		curriculums = []models.Curriculum{}
	}

	c.JSON(http.StatusOK, curriculums)
}

func GetAllCurriculums(c *gin.Context) {
	curriculums, err := services.GetAllCurriculums(c.Request.Context(), c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get curriculums", "code": "SERVER_ERROR"})
		return
	}

	if curriculums == nil {
		curriculums = []models.Curriculum{}
	}

	c.JSON(http.StatusOK, curriculums)
}

type UpdateCurriculumBody struct {
	Title   string `json:"title"`
	ClassID string `json:"class_id"`
}

func UpdateCurriculum(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}

	var body UpdateCurriculumBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "code": "INVALID_PAYLOAD"})
		return
	}

	updates := map[string]interface{}{
		"title":    body.Title,
		"class_id": body.ClassID,
	}

	if err := services.UpdateCurriculum(c.Request.Context(), id, updates, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update curriculum", "code": "SERVER_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteCurriculum(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required", "code": "ID_REQUIRED"})
		return
	}

	if err := services.DeleteCurriculum(c.Request.Context(), id, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete curriculum", "code": "OPERATION_FAILED"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
