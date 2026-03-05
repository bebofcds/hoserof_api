package services

import (
	"HOSEROF_API/config"
	"HOSEROF_API/middleware"
	"HOSEROF_API/models"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

func GetTeacherDashboard(c *gin.Context) (models.TeacherDashboardResponse, error) {

	ctx := c.Request.Context()
	services := config.GetServices(c)

	claims := c.MustGet("claims").(*middleware.Claims)
	classID := claims.UserClass

	studentsIter := services.Firebase.DB.
		Collection("students").
		Where("student_class", "==", classID).
		Documents(ctx)

	type studentInfo struct {
		ID   string
		Name string
	}

	var students []studentInfo

	for {
		doc, err := studentsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return models.TeacherDashboardResponse{}, err
		}

		var s models.UserFirestore
		if err := doc.DataTo(&s); err != nil {
			return models.TeacherDashboardResponse{}, err
		}

		students = append(students, studentInfo{
			ID:   doc.Ref.ID,
			Name: s.StudentName,
		})
	}

	totalStudents := len(students)

	// Create lookup map
	studentMap := make(map[string]string)
	for _, s := range students {
		studentMap[s.ID] = s.Name
	}

	iter := services.Firebase.DB.
		CollectionGroup("attendance").
		Documents(ctx)

	sessionMap := make(map[string][]bool)
	absenceCount := make(map[string]int)
	totalDays := make(map[string]int)

	for {

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return models.TeacherDashboardResponse{}, err
		}

		studentID := doc.Ref.Parent.Parent.ID

		// Skip students not in this class
		if _, ok := studentMap[studentID]; !ok {
			continue
		}

		var rec models.AttendanceRecord
		if err := doc.DataTo(&rec); err != nil {
			return models.TeacherDashboardResponse{}, err
		}

		date := doc.Ref.ID

		sessionMap[date] = append(sessionMap[date], rec.Attended)

		totalDays[studentID]++

		if !rec.Attended {
			absenceCount[studentID]++
		}
	}

	var dates []string
	for date := range sessionMap {
		dates = append(dates, date)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	var lastSession models.LastSessionStats

	if len(dates) > 0 {

		lastDate := dates[0]
		records := sessionMap[lastDate]

		attended := 0

		for _, a := range records {
			if a {
				attended++
			}
		}

		absent := len(records) - attended

		percentage := 0.0
		if len(records) > 0 {
			percentage = float64(attended) / float64(len(records)) * 100
		}

		parsedDate, _ := time.Parse("2006-01-02", lastDate)

		lastSession = models.LastSessionStats{
			SessionID:  lastDate,
			Date:       parsedDate,
			Attended:   attended,
			Absent:     absent,
			Percentage: percentage,
		}
	}

	var last4 []models.SessionChartItem

	for i := 0; i < len(dates) && i < 4; i++ {

		date := dates[i]
		records := sessionMap[date]

		attended := 0

		for _, a := range records {
			if a {
				attended++
			}
		}

		last4 = append(last4, models.SessionChartItem{
			Date:     date,
			Attended: attended,
			Absent:   len(records) - attended,
		})
	}

	var lowStudents []models.LowAttendanceStudent

	for _, s := range students {

		td := totalDays[s.ID]

		if td == 0 {
			continue
		}

		percentage := float64(td-absenceCount[s.ID]) / float64(td) * 100

		lowStudents = append(lowStudents, models.LowAttendanceStudent{
			ID:                   s.ID,
			Name:                 s.Name,
			AttendancePercentage: percentage,
		})
	}

	sort.Slice(lowStudents, func(i, j int) bool {
		return lowStudents[i].AttendancePercentage < lowStudents[j].AttendancePercentage
	})

	if len(lowStudents) > 5 {
		lowStudents = lowStudents[:5]
	}

	return models.TeacherDashboardResponse{
		TotalStudents:         totalStudents,
		LastSession:           lastSession,
		Last4Sessions:         last4,
		LowAttendanceStudents: lowStudents,
	}, nil
}

func GetAdminDashboard(c *gin.Context) (models.AdminDashboardResponse, error) {

	ctx := c.Request.Context()
	svcs := config.GetServices(c)

	var resp models.AdminDashboardResponse

	iter := svcs.Firebase.DB.Collection("students").Documents(ctx)

	totalStudents := 0
	todayAttendance := 0
	today := time.Now().Format("2006-01-02")

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return resp, err
		}

		totalStudents++

		var s models.UserFirestore
		if err := doc.DataTo(&s); err == nil {
			if s.LastAttendanceDate == today {
				todayAttendance++
			}
		}
	}

	examIter := svcs.Firebase.DB.
		Collection("exams").
		Where("status", "==", "pending").
		Documents(ctx)

	pendingExams := 0
	for {
		_, err := examIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return resp, err
		}
		pendingExams++
	}

	resultIter := svcs.Firebase.DB.
		Collection("results").
		Where("status", "==", "pending").
		Documents(ctx)

	pendingResults := 0
	for {
		_, err := resultIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return resp, err
		}
		pendingResults++
	}

	resp = models.AdminDashboardResponse{
		TotalStudents:   totalStudents,
		TodayAttendance: todayAttendance,
		PendingExams:    pendingExams,
		PendingResults:  pendingResults,
	}

	return resp, nil
}
