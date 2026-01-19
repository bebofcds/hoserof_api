package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/models"
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AttendanceService struct{}

func MarkAttendance(studentID string, attended bool) error {
	ctx := context.Background()

	studentDoc := config.DB.Collection("students").Doc(studentID)

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

func MarkAttendanceManual(studentID string, datetime string, attended bool) error {
	ctx := context.Background()

	studentDoc := config.DB.Collection("students").Doc(studentID)

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

func GetAttendance(studentID string) (models.AttendanceResponse, error) {
	ctx := context.Background()

	iter := config.DB.Collection("students").Doc(studentID).Collection("attendance").OrderBy("timestamp", firestore.Asc).Documents(ctx)

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
