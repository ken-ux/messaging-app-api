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
	for {
		conn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
		time.Sleep(time.Second * 3)
	}
}
