package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
		`SELECT description
		FROM settings
		LEFT JOIN "user" ON settings.user_id = "user".user_id
		WHERE username = '%s'`, user.Username,
	)).Scan(&profile.Description)
	if err != pgx.ErrNoRows {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("query error: %v", err))
		return
	}
	c.IndentedJSON(http.StatusOK, profile)
}
