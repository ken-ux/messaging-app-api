package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	"github.com/ken-ux/messaging-app-api/api"
	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/ws"
)

type User struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=20"`
}

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
	err := db.Init()
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
		os.Exit(1)
	}

	router := gin.Default()

	// Set-up CORS policy.
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("ORIGIN_URL")}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}
	router.Use(cors.New(config))

	router.GET("/", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, "Welcome to the backend.")
	})
	router.GET("/search", api.SearchUsers)
	router.POST("/auth", api.AuthenticateUser)
	router.POST("/auth/login", api.LoginUser)
	router.POST("/auth/register", api.RegisterUser)
	router.GET("/ws", ws.StartSocket)
	router.POST("/message", api.PostMessage)
	router.GET("/message", api.GetMessages)
	router.GET("/profile", api.GetProfile)

	router.Run()
}
