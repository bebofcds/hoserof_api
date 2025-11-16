package routes

import (
	"HOSEROF_API/controllers"
	"HOSEROF_API/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	protected.GET("/loginWithToken", controllers.TokenCheck)

	return r
}
