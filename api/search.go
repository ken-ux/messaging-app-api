package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/defs"
)

func SearchUsers(c *gin.Context) {
	var userList []defs.User
	username := c.Query("username")

	if username == "" {
		c.IndentedJSON(http.StatusOK, userList)
		return
	}

	rows, err := db.Pool.Query(context.Background(), fmt.Sprintf(
		`SELECT username
		FROM "user"
		WHERE username LIKE '%s%%'`,
		username,
	))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("query error: %v", err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user defs.User
		err := rows.Scan(&user.Username)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("query error: %v", err))
			return
		}
		userList = append(userList, user)
	}

	if err := rows.Err(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("query error: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, userList)
}
