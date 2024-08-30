package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ken-ux/messaging-app-api/db"
)

func SearchUsers(c *gin.Context) {
	username := c.Query("username")

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

	type User struct {
		username string
	}

	var userList []User

	for rows.Next() {
		var user User
		err := rows.Scan(user.username)
		if err != nil {
			log.Fatal(err)
		}
		userList = append(userList, user)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(userList)

	// return userList
}
