package routes

import (
	"HOSEROF_API/config"
	"HOSEROF_API/controllers"
	"HOSEROF_API/middleware"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter(svc *config.Services) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery(), middleware.BodySizeLimit)

	r.Use(func(c *gin.Context) {
		c.Set("services", svc)
		c.Next()
	})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API is running!",
		})
	})

	auth := r.Group("auth")
	auth.POST("/login", controllers.Login)

	admin := r.Group("/admin")
	admin.Use(middleware.RequireAuth(svc.JWTSecret), middleware.RequireAdmin())
	//Admin Users Management Endpoints
	users := admin.Group("/users")
	users.POST("/create-student", controllers.CreateStudent)
	users.POST("/create-staff", controllers.CreateStaff)
	users.PUT("/update-student/:studentID", controllers.UpdateStudent)
	users.DELETE("/delete-student/:studentID", controllers.DeleteStudent)
	users.PUT("/update-staff/:staffID", controllers.UpdateStaff)
	users.DELETE("/delete-staff/:staffID", controllers.DeleteStaff)
	users.GET("/student-profile/:userId", controllers.GetStudentByID)
	users.GET("/staff-profile/:userId", controllers.GetStaffByID)
	users.GET("/staff", controllers.GetStaff)
	//Admin Attendance Endpoints
	attendanceAdmin := admin.Group("/attendance")
	attendanceAdmin.POST("/mark-attendance", controllers.MarkAttendance)
	attendanceAdmin.GET("/get-attendance/:studentID", controllers.GetAttendanceByID)
	attendanceAdmin.POST("/mark-batch", controllers.MarkAttendanceBatch)
	attendanceAdmin.GET("/class/:classId", controllers.GetStudentsByClass)
	//Admin Exam Management Endpoints
	examAdmin := admin.Group("/exam")
	examAdmin.POST("/create-exam", controllers.CreateExam)
	examAdmin.DELETE("/delete-exam/:exam_id", controllers.DeleteExam)
	examAdmin.GET("/show-exam/:exam_id", controllers.GetExamForAdmin)
	examAdmin.GET("/submissions/:exam_id", controllers.GetSubmissionsForExam)
	examAdmin.POST("/release-results/:exam_id", controllers.ReleaseResults)
	examAdmin.GET("/list-exams", controllers.ListAllExams)
	examAdmin.GET("/submitted-exams/:student_id", controllers.GetStudentSubmittedExams)
	//Admin Curriculum Management Endpoints
	curriculumAdmin := admin.Group("/curriculum")
	curriculumAdmin.GET("/", controllers.GetAllCurriculums)
	curriculumAdmin.POST("/upload-curriculum", controllers.UploadCurriculum)
	curriculumAdmin.PUT("/:id", controllers.UpdateCurriculum)
	curriculumAdmin.DELETE("/:id", controllers.DeleteCurriculum)

	students := r.Group("/students")
	students.Use(middleware.RequireAuth(svc.JWTSecret))
	students.GET("/attendance/get-attendance", controllers.GetAttendance)
	exam := students.Group("/exam")
	exam.GET("/list-exams", controllers.ListExamsForStudent)
	exam.GET("/show-exam/:exam_id", controllers.GetExamForStudent)
	exam.POST("/submit-exam/:exam_id", controllers.SubmitExam)
	exam.GET("/get-result/:exam_id", controllers.GetReleasedResultForStudent)
	exam.GET("/list-results", controllers.ListReleasedResults)
	curriculum := students.Group("/curriculum")
	curriculum.GET("/class/:class_id", controllers.GetCurriculumsByClass)

	return r
}
