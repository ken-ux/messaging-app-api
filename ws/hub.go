package ws

import (
	"fmt"
	"net/http"

	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StartSocket(c *gin.Context) {
	// Don't do this, it's unwise to trust all origins.
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Do this instead to compare the request origin with the allowed origin.
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return r.Header["Origin"][0] == os.Getenv("ORIGIN_URL")
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Bad request: %v", err)
		return
	}
	defer conn.Close()

	c1 := make(chan string)

	go sendMessage(conn)
	go receiveMessage(conn, c1)

	for msg := range c1 {
		fmt.Println(msg)
	}
}

func receiveMessage(conn *websocket.Conn, c1 chan string) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			fmt.Printf("error: %v", err)
			break
		}
		c1 <- string(message)
	}
}

func sendMessage(conn *websocket.Conn) {
	for {
		// conn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
		time.Sleep(time.Second * 2)
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			break
		}
	}
}
