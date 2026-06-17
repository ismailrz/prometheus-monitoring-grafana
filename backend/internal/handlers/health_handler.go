package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func Ready(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "database unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
