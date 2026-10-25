package httpapi

import "github.com/gin-gonic/gin"

// fail writes the shared API error envelope.
func fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
