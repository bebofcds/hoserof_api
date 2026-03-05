package models

import "time"

type AttendanceRecord struct {
	Attended  bool      `json:"attended" firestore:"attended"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
}

type AttendanceResponse struct {
	Records    []AttendanceRecord `json:"records"`
	Percentage float64            `json:"percentage"`
}

type MarkAttendanceRequest struct {
	Attended bool `json:"attended"`
}
type AttendanceBatchRecord struct {
	StudentID string `json:"studentId"`
	Attended  bool   `json:"attended"`
}

type StudentAttendanceSummary struct {
	StudentID  string  `json:"studentId"`
	Name       string  `json:"name"`
	Class      string  `json:"class"`
	Grade      string  `json:"grade"`
	Phone      string  `json:"phone"`
	TotalDays  int     `json:"totalDays"`
	Attended   int     `json:"attendedDays"`
	Absent     int     `json:"absentDays"`
	Percentage float64 `json:"percentage"`
}
