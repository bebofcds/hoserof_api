package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodySizeLimit(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
	c.Next()
}
