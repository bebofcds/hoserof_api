package config

import (
	"github.com/gin-gonic/gin"
)

type Services struct {
	Firebase  *Firebase
	Supabase  *Supabase
	JWTSecret []byte
	JWT       interface{}
}

func GetServices(c *gin.Context) *Services {
	svc, _ := c.Get("services")
	return svc.(*Services)
}
