package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/defs"
)

func PostMessage(c *gin.Context) {
	var message defs.Message

	// // Bind JSON fields to form variable.
	if err := c.BindJSON(&message); err != nil {
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
		`WITH new_message AS (
			INSERT INTO "message" (author_id, message_body, creation_date)
			VALUES (
				(SELECT user_id FROM "user" WHERE username = '%s'),
				'%s',
				'%s'
			)
			RETURNING message_id
		)
		INSERT INTO recipient (user_id, message_id)
		VALUES (
			(SELECT user_id FROM "user" WHERE username = '%s'),
			(SELECT message_id from new_message)
		)`,
		message.Sender, message.Message_Body, time.Now().Format(time.RFC3339), message.Recipient))
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

func GetMessage(c *gin.Context) {
	fmt.Println("hello")
}
