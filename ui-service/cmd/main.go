package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static files (adjust the directory as needed)
	r.Static("/static", "./static")

	// Serve index.html for the root and all other GET requests (SPA fallback)
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.POST("/login", loginHandler)
	r.POST("/submit-sql", submitSQLHandler)

	r.Run(":8080")
}

func loginHandler(c *gin.Context) {
	// TODO: Implement JWT-based login
	c.JSON(http.StatusOK, gin.H{"token": "fake-jwt-token"})
}

func submitSQLHandler(c *gin.Context) {
	// TODO: Implement SQL submission to target DB
	c.JSON(http.StatusOK, gin.H{"result": "SQL submitted (stub)"})
}
