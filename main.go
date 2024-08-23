package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password" validate:"required,min=5,max=20"`
}

var dbpool *pgxpool.Pool

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
	var dbpool_err error
	dbpool, dbpool_err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if dbpool_err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", dbpool_err)
		os.Exit(1)
	}
	defer dbpool.Close()

	router := gin.Default()

	// Allow all origins.
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
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
		errors := err.(validator.ValidationErrors)
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Validation error: %v", errors))
		return
	}

	// Check that username isn't taken.
	err := dbpool.QueryRow(context.Background(), fmt.Sprintf(
		`SELECT username
		FROM "user"
		WHERE username = '%s'`,
		user.Username)).
		Scan(&user.Username)

	if err == nil {
		c.IndentedJSON(http.StatusBadRequest, "Username already exists.")
		return
	}

	// Hash password and re-bind to user variable.
	hash, err := HashPassword(user.Password)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Encryption error: %v", err))
		return
	}
	user.Password = hash

	// Add user to database.
	// Begin transaction.
	tx, err := dbpool.Begin(context.Background())
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Error: %v", err))
		return
	}

	// Rollback transaction if it doesn't commit successfully.
	defer tx.Rollback(context.Background())

	// Execute insert statement.
	_, err = tx.Exec(context.Background(), fmt.Sprintf(
		`INSERT INTO "user" (username, password)
		VALUES (
			'%s',
			'%s'
		)`,
		user.Username, user.Password))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Commit transaction.
	err = tx.Commit(context.Background())
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Send JWT to client.
	signedToken, err := GetToken(user, c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, signedToken)
}

func loginUser(c *gin.Context) {
	var user User

	// Bind JSON fields from form data.
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Check if user exists in database.
	err := dbpool.QueryRow(context.Background(), fmt.Sprintf(
		`SELECT username
		FROM "user"
		WHERE username = '%s'`,
		user.Username)).
		Scan(&user.Username)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "User doesn't exist.")
		return
	}

	// Check if password hash matches.
	var hashedPassword string

	err = dbpool.QueryRow(context.Background(), fmt.Sprintf(
		`SELECT password
		FROM "user"
		WHERE username = '%s'`,
		user.Username)).
		Scan(&hashedPassword)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "Error fetching password.")
		return
	}

	match := CheckPasswordHash(user.Password, hashedPassword)

	if !match {
		c.IndentedJSON(http.StatusBadRequest, "Wrong password.")
		return
	}

	// Send JWT to client.
	signedToken, err := GetToken(user, c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, signedToken)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetToken(user User, c *gin.Context) (signedToken string, err error) {
	// Load secret and cast from string to []byte.
	secret := os.Getenv("SECRET")
	key, err := base64.RawStdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("JWT error: %v", err)
	}

	// Encode user-specific information into token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": user.Username,
			"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		})

	// Sign token with key.
	signedToken, err = token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("JWT error: %v", err)
	}
	return
}

func ValidateToken(tokenString string, c *gin.Context) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate for expected alg aka signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get secret and convert to []byte.
		secret := os.Getenv("SECRET")
		key, err := base64.RawStdEncoding.DecodeString(secret)
		if err != nil {
			return nil, fmt.Errorf("error parsing secret: %v", err)
		}
		return key, nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// if claims, ok := token.Claims.(jwt.MapClaims); ok {
	// 	fmt.Println(claims["foo"], claims["nbf"])
	// } else {
	// 	fmt.Println(err)
	// }
	return true
}
