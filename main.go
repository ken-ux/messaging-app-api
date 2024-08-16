package main

import (
	// "context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// "github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type NewUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// var dbpool *pgxpool.Pool

func main() {
	// Load dev environment.
	env := os.Getenv("ENV_NAME")
	if env != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Connect to database.
	// var dbpool_err error
	// dbpool, dbpool_err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	// if dbpool_err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", dbpool_err)
	// 	os.Exit(1)
	// }
	// defer dbpool.Close()

	router := gin.Default()

	// Allow all origins.
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		fmt.Println("Hello, World!")
		c.IndentedJSON(http.StatusOK, "Welcome to the backend.")
	})
	router.POST("/login", loginUser)
	router.POST("/register", registerUser)

	router.Run()
}

func registerUser(c *gin.Context) {
	var user NewUser

	// Bind JSON fields from form data.
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Validate form inputs for invalid characters and check if username is taken.

	// Bind hashed password to user variable.

	// Send query to backend to register user.

	// Send JWT to client.
}

func loginUser(c *gin.Context) {
}
