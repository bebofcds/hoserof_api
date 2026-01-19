package routes

import (
	"HOSEROF_API/controllers"
	"HOSEROF_API/middleware"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API is running!",
		})
	})

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	protected.GET("/loginWithToken", controllers.TokenCheck)

	attendance := r.Group("/attendance")
	attendance.Use(middleware.RequireAuth())
	attendance.GET("/get", controllers.GetAttendance)

	attendanceAdmin := attendance.Group("/")
	attendanceAdmin.Use(middleware.RequireAuth())
	attendanceAdmin.Use(middleware.RequireAdmin())
	attendanceAdmin.POST("/mark", controllers.MarkAttendance)
	attendanceAdmin.GET("/get/:studentID", controllers.GetAttendanceByID)
	attendanceAdmin.POST("/manual", controllers.MarkAttendanceManual)

	exam := r.Group("/exam")
	exam.Use(middleware.RequireAuth())
	exam.GET("/list", controllers.ListExamsForStudent)
	exam.GET("/:examID", controllers.GetExamForStudent)
	exam.POST("/submit/:examID", controllers.SubmitExam)
	exam.GET("/result/:examID", controllers.GetReleasedResultForStudent)
	exam.GET("/results", controllers.ListReleasedResults)

	examAdmin := exam.Group("/")
	examAdmin.Use(middleware.RequireAuth())
	examAdmin.Use(middleware.RequireAdmin())
	examAdmin.POST("/create", controllers.CreateExam)
	examAdmin.GET("/submissions/:examID", controllers.GetSubmissionsForExam)
	examAdmin.POST("/grade/:examID", controllers.GradeAnswer)
	examAdmin.POST("/release/:examID", controllers.ReleaseResultsHandler)

	curriculum := r.Group("/curriculum")
	curriculum.Use(middleware.RequireAuth())
	curriculum.GET("/class/:class_id", controllers.GetCurriculumsByClass)

	curriculumAdmin := curriculum.Group("/")
	curriculumAdmin.Use(middleware.RequireAuth())
	curriculumAdmin.Use(middleware.RequireAdmin())
	curriculumAdmin.GET("/", controllers.GetAllCurriculums)
	curriculumAdmin.POST("/upload", controllers.UploadCurriculum)
	curriculumAdmin.PUT("/:id", controllers.UpdateCurriculum)
	curriculumAdmin.DELETE("/:id", controllers.DeleteCurriculum)

	return r
}
