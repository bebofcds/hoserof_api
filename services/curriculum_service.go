/*
================================================================================
HOSEROF_API - Curriculum Services
================================================================================

Description:
This package provides service-level functions for managing curriculum files
within the HosErof system. It supports uploading, retrieving, updating,
and deleting curriculum files, while storing metadata in Firestore and files
in Supabase storage.

Responsibilities:
1. UploadCurriculum        - Uploads a curriculum file to Supabase, saves metadata in Firestore.
2. GetCurriculumsByClass   - Retrieves all curriculums for a specific class.
3. GetAllCurriculums       - Retrieves all curriculums in the system, sorted by class.
4. UpdateCurriculum        - Updates curriculum metadata (title, class_id).
5. DeleteCurriculum        - Deletes a curriculum file from Supabase and its Firestore metadata.
6. getFileType             - Helper function to classify file type based on extension.

Usage Notes:
- All functions require a Gin context (`*gin.Context`) to extract services.
- UploadCurriculum requires multipart file input.
- File types are automatically inferred for categorization.
- Firestore collections:
    - "curriculums" → stores metadata for each curriculum.
- Supabase storage bucket: "Curriculum".

Error Handling:
- UploadCurriculum removes uploaded file from Supabase if Firestore save fails.
- DeleteCurriculum logs a warning if the file cannot be removed from Supabase, but continues.
- Iteration over Firestore documents handles `iterator.Done` gracefully.

Date Format Reference:
- CreatedAt and UpdatedAt timestamps use `time.Now()` in Go's local timezone.
================================================================================
*/

package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// UploadCurriculum uploads a file to Supabase, stores metadata in Firestore, and returns the curriculum.
func UploadCurriculum(ctx context.Context, req models.UploadCurriculumRequest, file multipart.File, header *multipart.FileHeader, userID string, c *gin.Context) (*models.Curriculum, error) {
	ext := filepath.Ext(header.Filename)
	services := config.GetServices(c)

	uniqueFilename := fmt.Sprintf("%s_%s%s", req.ClassID, uuid.New().String(), ext)
	storagePath := fmt.Sprintf("%s/%s", req.ClassID, uniqueFilename)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileReader := bytes.NewReader(fileBytes)

	// Upload file to Supabase
	_, err = services.Supabase.Storage.UploadFile("Curriculum", storagePath, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to supabase: %w", err)
	}
	// Get public URL of uploaded file
	resp := services.Supabase.Storage.GetPublicUrl("Curriculum", storagePath)
	fileURL := resp.SignedURL

	fileType := getFileType(ext)

	curriculum := &models.Curriculum{
		ID:        uuid.New().String(),
		ClassID:   req.ClassID,
		Title:     req.Title,
		FileType:  fileType,
		FileURL:   fileURL,
		FileName:  header.Filename,
		FileSize:  header.Size,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Save metadata in Firestore
	_, err = services.Firebase.DB.Collection("curriculums").Doc(curriculum.ID).Set(ctx, curriculum)
	if err != nil {
		// Remove file from Supabase if Firestore save fails
		services.Supabase.Storage.RemoveFile("Curriculum", []string{storagePath})
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return curriculum, nil
}

// GetCurriculumsByClass returns all curriculums for a specific class.
func GetCurriculumsByClass(ctx context.Context, classID string, c *gin.Context) ([]models.Curriculum, error) {
	var curriculums []models.Curriculum
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("curriculums").
		Where("class_id", "==", classID).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var curriculum models.Curriculum
		if err := doc.DataTo(&curriculum); err != nil {
			return nil, fmt.Errorf("failed to parse curriculum: %w", err)
		}
		curriculums = append(curriculums, curriculum)
	}

	return curriculums, nil
}

// GetAllCurriculums returns all curriculums sorted by class ID.
func GetAllCurriculums(ctx context.Context, c *gin.Context) ([]models.Curriculum, error) {
	var curriculums []models.Curriculum
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("curriculums").
		OrderBy("class_id", firestore.Asc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var curriculum models.Curriculum
		if err := doc.DataTo(&curriculum); err != nil {
			return nil, fmt.Errorf("failed to parse curriculum: %w", err)
		}
		curriculums = append(curriculums, curriculum)
	}

	return curriculums, nil
}

// UpdateCurriculum updates curriculum metadata (title, class ID) and timestamps.
func UpdateCurriculum(ctx context.Context, id string, updates map[string]interface{}, c *gin.Context) error {
	updates["updated_at"] = time.Now()
	services := config.GetServices(c)

	_, err := services.Firebase.DB.Collection("curriculums").Doc(id).Update(ctx, []firestore.Update{
		{Path: "title", Value: updates["title"]},
		{Path: "class_id", Value: updates["class_id"]},
		{Path: "updated_at", Value: updates["updated_at"]},
	})

	return err
}

// DeleteCurriculum removes a curriculum file from Supabase and its metadata from Firestore.
func DeleteCurriculum(ctx context.Context, id string, c *gin.Context) error {
	services := config.GetServices(c)

	doc, err := services.Firebase.DB.Collection("curriculums").Doc(id).Get(ctx)
	if err != nil {
		return fmt.Errorf("curriculum not found: %w", err)
	}

	var curriculum models.Curriculum
	if err := doc.DataTo(&curriculum); err != nil {
		return fmt.Errorf("failed to parse curriculum: %w", err)
	}

	ext := filepath.Ext(curriculum.FileName)
	filenameFromID := fmt.Sprintf("%s_%s%s", curriculum.ClassID, id, ext)
	storagePath := fmt.Sprintf("curriculum/%s/%s", curriculum.ClassID, filenameFromID)

	// Remove file from Supabase
	if _, err = services.Supabase.Storage.RemoveFile("curriculum", []string{storagePath}); err != nil {
		fmt.Printf("Warning: failed to delete file from Supabase: %v\n", err)
	}
	// Remove metadata from Firestore
	if _, err = services.Firebase.DB.Collection("curriculums").Doc(id).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from firestore: %w", err)
	}
	return nil
}

// getFileType returns the type of file based on its extension.
func getFileType(ext string) string {
	switch ext {
	case ".pdf":
		return "pdf"
	case ".mp3", ".wav", ".ogg", ".m4a", ".mpeg":
		return "audio"
	case ".mp4", ".avi", ".mov", ".wmv":
		return "video"
	case ".doc", ".docx", ".txt", ".rtf":
		return "document"
	case ".ppt", ".pptx":
		return "presentation"
	case ".xls", ".xlsx":
		return "spreadsheet"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	default:
		return "other"
	}
}
