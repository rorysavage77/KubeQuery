package main

import (
	"net/http"

	"mime"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ensure .js files are served with the correct MIME type
	mime.AddExtensionType(".js", "application/javascript")

	r := gin.Default()

	// Allow CORS for local dev
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Error string `json:"error,omitempty"`
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Error: "Invalid request"})
		return
	}
	// TODO: Replace with real authentication
	if req.Username == "admin" && req.Password == "admin" {
		c.JSON(http.StatusOK, LoginResponse{Token: "fake-jwt-token"})
	} else {
		c.JSON(http.StatusUnauthorized, LoginResponse{Error: "Invalid credentials"})
	}
}

type SQLRequest struct {
	SQL string `json:"sql"`
}

type SQLResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func submitSQLHandler(c *gin.Context) {
	// TODO: Validate JWT from Authorization header
	var req SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SQLResponse{Error: "Invalid request"})
		return
	}
	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, SQLResponse{Error: "SQL is required"})
		return
	}
	// TODO: Integrate with KubeQuery controller or AI
	c.JSON(http.StatusOK, SQLResponse{Result: "SQL executed (stub): " + req.SQL})
}
