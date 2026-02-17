/*
================================================================================
HOSEROF_API - Attendance Service Package
================================================================================

Description:
This package contains service-level functions for managing student attendance
and user data. All functions interact primarily with Firebase Firestore, using
the injected services from the context (config.Services).

Responsibilities:
1. Mark attendance for a student (automatic for today or manual for a specific datetime).
2. Retrieve attendance records and calculate attendance percentage.
3. Retrieve students by class, with optional filtering for already marked attendance.
4. Batch mark attendance for multiple students.
5. Retrieve user data by student ID.

Usage Notes:
- All functions require a Gin context (`*gin.Context`) to extract services and
  the request context.
- Time-sensitive operations use `time.Now()`; manual entries require specific
  datetime format: "2006-01-02;15:04:05".
- Firebase Firestore collections:
  - `students` → top-level student documents
  - `attendance` → sub-collection under each student document

Error Handling:
- If a student or document is not found, functions return an appropriate error.
- Firestore errors are returned directly to the caller for logging or response.

Important:
- Ensure the services are properly injected into Gin context via `config.GetServices(c)`.
- Batch operations use Firestore transactions for atomic writes.
- Attendance percentage is calculated as (attended sessions / total sessions) * 100.

Date Format Reference:
- Default daily attendance: "YYYY-MM-DD" (e.g., "2026-02-16")
- Manual attendance: "YYYY-MM-DD;HH:MM:SS" (e.g., "2026-02-16;14:30:00")

================================================================================
*/

package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MarkAttendance marks attendance for today for a specific student.
func MarkAttendance(studentID string, attended bool, c *gin.Context) error {
	ctx := c.Request.Context()
	services := config.GetServices(c)
	studentDoc := services.Firebase.DB.Collection("students").Doc(studentID)

	snap, err := studentDoc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.New("no user found")
		}
		return err
	}

	if !snap.Exists() {
		return errors.New("no user found")
	}

	attendanceDoc := studentDoc.
		Collection("attendance").
		Doc(time.Now().Format("2006-01-02"))

	_, err = attendanceDoc.Set(ctx, map[string]interface{}{
		"attended":  attended,
		"timestamp": firestore.ServerTimestamp,
	}, firestore.MergeAll)

	return err
}

// MarkAttendanceManual marks attendance for a specific datetime.
func MarkAttendanceManual(studentID string, datetime string, attended bool, c *gin.Context) error {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	studentDoc := services.Firebase.DB.Collection("students").Doc(studentID)
	snap, err := studentDoc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.New("no user found")
		}
		return err
	}

	if !snap.Exists() {
		return errors.New("no user found")
	}

	parsedTime, err := time.Parse("2006-01-02;15:04:05", datetime)
	if err != nil {
		return errors.New("invalid datetime format: expected 2006-01-02;15:04:05")
	}
	dateStr := parsedTime.Format("2006-01-02")
	attendanceDoc := studentDoc.Collection("attendance").Doc(dateStr)

	_, err = attendanceDoc.Set(ctx, map[string]interface{}{
		"attended":  attended,
		"timestamp": parsedTime,
	})

	return err
}

// GetAttendance retrieves all attendance records for a student and calculates percentage.
func GetAttendance(studentID string, c *gin.Context) (models.AttendanceResponse, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("students").Doc(studentID).Collection("attendance").OrderBy("timestamp", firestore.Asc).Documents(ctx)

	var records []models.AttendanceRecord

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return models.AttendanceResponse{}, err
		}
		var rec models.AttendanceRecord
		if err := doc.DataTo(&rec); err != nil {
			return models.AttendanceResponse{}, err
		}
		records = append(records, rec)
	}
	total := len(records)
	if total == 0 {
		return models.AttendanceResponse{
			Records:    []models.AttendanceRecord{},
			Percentage: 0,
		}, nil
	}

	attendedCount := 0

	for _, r := range records {
		if r.Attended {
			attendedCount++
		}
	}
	percentage := float64(attendedCount) / float64(total) * 100

	return models.AttendanceResponse{
		Records:    records,
		Percentage: percentage,
	}, nil

}

// GetStudentsByClass retrieves students in a specific class with optional hiding of already marked attendance.
func GetStudentsByClass(classID string, hideMarked bool, c *gin.Context) ([]models.UserClassList, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	iter := services.Firebase.DB.Collection("students").
		Where("student_class", "==", classID).
		Documents(ctx)

	var students []models.UserClassList
	today := time.Now().Format("2006-01-02")

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var s models.UserClassList
		if err := doc.DataTo(&s); err != nil {
			return nil, err
		}
		s.StudentID = doc.Ref.ID

		if !hideMarked {
			students = append(students, s)
			continue
		}

		attendanceDoc := services.Firebase.DB.
			Collection("students").
			Doc(s.StudentID).
			Collection("attendance").
			Doc(today)

		_, err = attendanceDoc.Get(ctx)
		// Already marked today, skip
		if err == nil {
			continue
		}

		if status.Code(err) == codes.NotFound {
			// Not marked yet, include
			students = append(students, s)
			continue
		}

		return nil, err

	}

	return students, nil
}

// MarkAttendanceBatch marks attendance for multiple students in a batch.
func MarkAttendanceBatch(records []struct {
	StudentID string `json:"studentId"`
	Attended  bool   `json:"attended"`
}, c *gin.Context) error {
	services := config.GetServices(c)

	ctx := c.Request.Context()
	batch := services.Firebase.DB.Batch()
	today := time.Now().Format("2006-01-02")

	for _, r := range records {

		if r.StudentID == "" {
			continue
		}

		attendanceRef := services.Firebase.DB.
			Collection("students").
			Doc(r.StudentID).
			Collection("attendance").
			Doc(today)

		batch.Set(attendanceRef, map[string]interface{}{
			"attended":  r.Attended,
			"timestamp": firestore.ServerTimestamp,
		}, firestore.MergeAll)
	}

	_, err := batch.Commit(ctx)
	return err
}

// GetUserByID retrieves a single user by their Firestore ID.
func GetUserByID(userID string, c *gin.Context) (models.UserFirestore, error) {
	ctx := c.Request.Context()
	services := config.GetServices(c)

	doc, err := services.Firebase.DB.Collection("students").Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.UserFirestore{}, errors.New("user not found")
		}
		return models.UserFirestore{}, err
	}

	var user models.UserFirestore
	if err := doc.DataTo(&user); err != nil {
		return models.UserFirestore{}, err
	}

	user.StudentID = doc.Ref.ID
	return user, nil
}
