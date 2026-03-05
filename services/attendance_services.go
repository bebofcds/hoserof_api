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
	"context"
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
	svcs := config.GetServices(c)

	return svcs.Firebase.DB.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		studentRef := svcs.Firebase.DB.Collection("students").Doc(studentID)

		_, err := tx.Get(studentRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return errors.New("no user found")
			}
			return err
		}

		today := time.Now().Format("2006-01-02")
		attendanceRef := studentRef.Collection("attendance").Doc(today)

		_, err = tx.Get(attendanceRef)
		alreadyMarked := err == nil

		// Write attendance record
		if err := tx.Set(attendanceRef, map[string]interface{}{
			"attended":  attended,
			"timestamp": firestore.ServerTimestamp,
		}, firestore.MergeAll); err != nil {
			return err
		}

		// Only update counters if not already marked
		if !alreadyMarked {

			update := map[string]interface{}{
				"total_days":           firestore.Increment(1),
				"last_attendance_date": today,
			}

			if attended {
				update["attended_days"] = firestore.Increment(1)
			} else {
				update["absent_days"] = firestore.Increment(1)
			}

			if err := tx.Set(studentRef, update, firestore.MergeAll); err != nil {
				return err
			}
		}

		return nil
	})
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

// GetStudents retrieves students in a specific class with optional hiding of already marked attendance.
func GetStudents(classID string, hideMarked bool, c *gin.Context) ([]models.UserClassList, error) {

	ctx := c.Request.Context()
	svcs := config.GetServices(c)

	query := svcs.Firebase.DB.Collection("students").
		Where("student_class", "==", classID)

	if hideMarked {
		today := time.Now().Format("2006-01-02")
		query = query.Where("last_attendance_date", "!=", today)
	}

	iter := query.Documents(ctx)

	students := []models.UserClassList{}

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
		students = append(students, s)
	}

	return students, nil
}

// MarkAttendanceBatch marks attendance for multiple students in a batch.
func MarkAttendanceBatch(records []models.AttendanceBatchRecord, c *gin.Context) error {

	ctx := c.Request.Context()
	svcs := config.GetServices(c)

	return svcs.Firebase.DB.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		today := time.Now().Format("2006-01-02")

		type studentState struct {
			ref           *firestore.DocumentRef
			attendanceRef *firestore.DocumentRef
			alreadyMarked bool
			attended      bool
		}

		states := make([]studentState, 0, len(records))

		for _, r := range records {

			if r.StudentID == "" {
				continue
			}

			studentRef := svcs.Firebase.DB.Collection("students").Doc(r.StudentID)
			attendanceRef := studentRef.Collection("attendance").Doc(today)

			_, err := tx.Get(attendanceRef)

			alreadyMarked := err == nil

			states = append(states, studentState{
				ref:           studentRef,
				attendanceRef: attendanceRef,
				alreadyMarked: alreadyMarked,
				attended:      r.Attended,
			})
		}

		for _, s := range states {

			tx.Set(s.attendanceRef, map[string]interface{}{
				"attended":  s.attended,
				"timestamp": firestore.ServerTimestamp,
			}, firestore.MergeAll)

			if !s.alreadyMarked {

				update := map[string]interface{}{
					"total_days":           firestore.Increment(1),
					"last_attendance_date": today,
				}

				if s.attended {
					update["attended_days"] = firestore.Increment(1)
				} else {
					update["absent_days"] = firestore.Increment(1)
				}

				tx.Set(s.ref, update, firestore.MergeAll)
			}
		}

		return nil
	})
}
func GetClassAttendanceSummary(classID string, c *gin.Context) ([]models.StudentAttendanceSummary, error) {
	ctx := c.Request.Context()
	svcs := config.GetServices(c)

	iter := svcs.Firebase.DB.Collection("students").
		Where("student_class", "==", classID).
		Documents(ctx)

	var result []models.StudentAttendanceSummary

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var student models.UserFirestore
		if err := doc.DataTo(&student); err != nil {
			return nil, err
		}

		percentage := 0.0
		if student.TotalDays > 0 {
			percentage = float64(student.AttendedDays) / float64(student.TotalDays) * 100
		}

		result = append(result, models.StudentAttendanceSummary{
			StudentID:  doc.Ref.ID,
			Name:       student.StudentName,
			Class:      student.StudentClass,
			Grade:      student.StudentGrade,
			Phone:      student.StudentPhone,
			TotalDays:  student.TotalDays,
			Attended:   student.AttendedDays,
			Absent:     student.AbsentDays,
			Percentage: percentage,
		})
	}

	return result, nil
}
