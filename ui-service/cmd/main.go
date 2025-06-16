package main

import (
	"database/sql"
	"fmt"
	"mime"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Ensure .js files are served with the correct MIME type
	mime.AddExtensionType(".js", "application/javascript")

	// Read DB config from env
	pgHost := os.Getenv("PGHOST")
	pgPort := os.Getenv("PGPORT")
	pgUser := os.Getenv("PGUSER")
	pgPass := os.Getenv("PGPASSWORD")
	pgDB := os.Getenv("PGDATABASE")
	if pgHost == "" {
		pgHost = "kubequery-postgres"
	}
	if pgPort == "" {
		pgPort = "5432"
	}
	if pgUser == "" {
		pgUser = "kquser"
	}
	if pgPass == "" {
		pgPass = "changeme"
	}
	if pgDB == "" {
		pgDB = "kqdb"
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPass, pgDB)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to open DB: %v", err))
	}
	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}

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
	Rows  []map[string]interface{} `json:"rows,omitempty"`
	Error string                   `json:"error,omitempty"`
}

func submitSQLHandler(c *gin.Context) {
	var req SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SQLResponse{Error: "Invalid request"})
		return
	}
	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, SQLResponse{Error: "SQL is required"})
		return
	}
	rows, err := db.Query(req.SQL)
	if err != nil {
		c.JSON(http.StatusBadRequest, SQLResponse{Error: err.Error()})
		return
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, SQLResponse{Error: err.Error()})
		return
	}
	results := []map[string]interface{}{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			c.JSON(http.StatusInternalServerError, SQLResponse{Error: err.Error()})
			return
		}
		rowMap := map[string]interface{}{}
		for i, col := range cols {
			v := vals[i]
			b, ok := v.([]byte)
			if ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = v
			}
		}
		results = append(results, rowMap)
	}
	c.JSON(http.StatusOK, SQLResponse{Rows: results})
}
