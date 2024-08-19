package main

import (
	// "context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// "github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type User struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=20"`
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
	var user User

	// Bind JSON fields from form data.
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Validate form input.
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(user); err != nil {
		// Validation failed, handle the error
		errors := err.(validator.ValidationErrors)
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Validation error: %v", errors))
		return
	}

	fmt.Println(user)

	// Bind hashed password to user variable.

	fmt.Println(user)

	// Send query to backend to register user.

	// Send JWT to client.
	// Load secret and cast from string to []byte.
	secret := os.Getenv("SECRET")
	key, err := base64.RawStdEncoding.DecodeString(secret)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
	}

	// Encode user-specific information into token.
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": user.Username,
		})

	// Sign token with key.
	s, err := t.SignedString(key)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
	}

	c.IndentedJSON(http.StatusOK, s)
}

func loginUser(c *gin.Context) {
}
