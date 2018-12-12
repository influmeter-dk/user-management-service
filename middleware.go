package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func bindUserFromBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User

		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payload missing"})
			c.Abort()
			return
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if user.Role == "" {
			user.Role = "PARTICIPANT"
		}

		c.Set("user", user)

		c.Next()
	}
}
