package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"

	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/defs"
)

func AuthenticateUser(c *gin.Context) {
	var user defs.User

	// Bind JSON fields from form data.
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	reqToken := c.Request.Header["Authorization"]
	splitToken := strings.Split(reqToken[0], "Bearer ")
	if len(splitToken) != 2 {
		c.IndentedJSON(http.StatusBadRequest, "Bad request: Bearer token not in proper format")
		return
	}
	token := splitToken[1]

	verified, err := ValidateToken(user, token)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, verified)
}

func RegisterUser(c *gin.Context) {
	var user defs.User

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
	err := db.Pool.QueryRow(context.Background(), fmt.Sprintf(
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
	tx, err := db.Pool.Begin(context.Background())
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
	signedToken, err := GetToken(user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, signedToken)
}

func LoginUser(c *gin.Context) {
	var user defs.User

	// Bind JSON fields from form data.
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Check if user exists in database.
	err := db.Pool.QueryRow(context.Background(), fmt.Sprintf(
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

	err = db.Pool.QueryRow(context.Background(), fmt.Sprintf(
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
	signedToken, err := GetToken(user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("JWT error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, signedToken)
}

func GetToken(user defs.User) (signedToken string, err error) {
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

func ValidateToken(user defs.User, tokenString string) (valid bool, err error) {
	// Parse token.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Confirm expected alg aka signing method.
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
		return false, fmt.Errorf("token parsing issue: %v", err)
	}

	// Extract claims from token.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, fmt.Errorf("issue mapping jwt claims")
	}

	// Check claims.
	if claims["sub"] != user.Username {
		return false, fmt.Errorf("sub doesn't match token")
	}

	expiry, err := claims.GetExpirationTime()
	if err != nil {
		return false, fmt.Errorf("invalid expiry")
	}

	if time.Now().After(expiry.Time) {
		return false, fmt.Errorf("token expired")
	}

	return true, nil
}
