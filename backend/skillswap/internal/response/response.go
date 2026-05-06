package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalError logs the error server-side and sends a generic 500 response.
// Internal error details are never exposed to the client.
func InternalError(c *gin.Context, err error) {
	log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}