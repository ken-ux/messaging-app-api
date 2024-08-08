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

	router.Run()
}

// func getUser(c *gin.Context) {
// }
