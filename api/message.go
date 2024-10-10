package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ken-ux/messaging-app-api/db"
	"github.com/ken-ux/messaging-app-api/defs"
)

func PostMessage(c *gin.Context) {
	var message defs.Message

	// // Bind JSON fields to message variable.
	if err := c.BindJSON(&message); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	// Append apostrophes with extra apostrophe, otherwise throws SQL error.
	message.Message_Body = strings.Replace(message.Message_Body, "'", "''", -1)

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

func GetMessages(c *gin.Context) {
	sender := c.Query("sender")
	recipient := c.Query("recipient")
	var user defs.User
	var messages []defs.Message

	user.Username = sender

	// Authenticate request first
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

	if err := queryMessages(&messages, sender, recipient); err != nil {
		c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", err))
		return
	}

	c.IndentedJSON(http.StatusOK, messages)
}

func queryMessages(messages *[]defs.Message, sender string, recipient string) (err error) {
	// Query database
	rows, err := db.Pool.Query(context.Background(), fmt.Sprintf(
		`SELECT sender, recipient, message_body, creation_date
		FROM
			(SELECT t1.username AS sender, t2.username AS recipient, message_body, creation_date
			FROM
				(SELECT message_id, username, message_body, creation_date
				FROM message
				INNER JOIN "user" ON message.author_id = "user".user_id
				WHERE username = '%s'
				ORDER BY creation_date ASC) t1
			INNER JOIN
				(SELECT username, message_id
				FROM recipient
				INNER JOIN "user" ON recipient.user_id = "user".user_id
				WHERE username = '%s') t2
			ON t1.message_id = t2.message_id) tbl_1
		UNION
			(SELECT t1.username AS sender, t2.username AS recipient, message_body, creation_date
			FROM
				(SELECT message_id, username, message_body, creation_date
				FROM message
				INNER JOIN "user" ON message.author_id = "user".user_id
				WHERE username = '%s'
				ORDER BY creation_date ASC) t1
			INNER JOIN
				(SELECT username, message_id
				FROM recipient
				INNER JOIN "user" on recipient.user_id = "user".user_id
				WHERE username = '%s') t2
			ON t1.message_id = t2.message_id)
		ORDER BY creation_date DESC
		LIMIT 15`, sender, recipient, recipient, sender))
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	// Releases any resources held by the rows no matter how the function returns.
	// Looping all the way through the rows also closes it implicitly,
	// but it is better to use defer to make sure rows is closed no matter what.
	defer rows.Close()

	// Iterate through all rows returned from the query.
	for rows.Next() {
		// Loop through rows, using Scan to assign column data to struct fields.
		var message defs.Message
		if err := rows.Scan(&message.Sender, &message.Recipient, &message.Message_Body, &message.Creation_Date); err != nil {
			return fmt.Errorf("query failed: %v", err)
		}
		*messages = append(*messages, message)
	}

	// Check if there were any issues when reading rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("query failed: %v", err)
	}
	return nil
}
