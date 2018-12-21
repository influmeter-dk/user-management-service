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
		if user.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing email/username"})
			c.Abort()
			return
		}
		if user.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing password"})
			c.Abort()
			return
		}

		u := User{
			Email:    user.Email,
			Password: user.Password,
			Role:     "PARTICIPANT",
		}

		c.Set("user", u)

		c.Next()
	}
}
