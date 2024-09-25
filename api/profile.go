package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/defs"
)

func GetProfile(c *gin.Context) {
	user := defs.User{Username: c.Query("username")}
	var profile defs.Profile

	if user.Username == "" {
		c.IndentedJSON(http.StatusBadRequest, "username cannot be blank")
		return
	}

	err := db.Pool.QueryRow(context.Background(), fmt.Sprintf(
		`SELECT description, CONVERT_FROM(color, 'utf-8')
		FROM settings
		LEFT JOIN "user" ON settings.user_id = "user".user_id
		WHERE username = '%s'`, user.Username,
	)).Scan(&profile.Description, &profile.Color)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("query error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, profile)
}

func UpdateProfile(c *gin.Context) {
	user := defs.User{Username: c.Query("username")}
	var profile defs.Profile

	if user.Username == "" {
		c.IndentedJSON(http.StatusBadRequest, "username cannot be blank")
		return
	}

	if err := c.BindJSON(&profile); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Authenticate request first.
	token, err := ParseTokenFromHeader(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	_, err = ValidateToken(user, token)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

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
		`INSERT INTO settings (user_id, description, color)
		VALUES (
			(SELECT user_id FROM "user" WHERE username = '%s'),
			'%s',
			'%s'::bytea
		)
		ON CONFLICT(user_id)
		DO UPDATE SET
			description = EXCLUDED.description,
			color = EXCLUDED.color`,
		user.Username, profile.Description, profile.Color))
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

	c.IndentedJSON(http.StatusOK, "ok")
}
