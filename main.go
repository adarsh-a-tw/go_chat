package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {

	router := gin.Default()
	router.StaticFile("/", "./static/index.html")
	router.GET("/ws", serveWs)
	err := router.Run()
	if err != nil {
		log.Fatalf("Unable to start server. Error %v", err)
	}
	log.Println("Server started successfully.")
}

func serveWs(c *gin.Context) {

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error in upgrading web socket. Error: %v", err)
		return
	}

	go handleClient(conn)
}

var clients = make(map[*websocket.Conn]struct{})

type Message struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

func broadcast(msg Message) {
	for conn := range clients {
		conn.WriteJSON(msg)
	}
}

func handleClient(c *websocket.Conn) {
	defer func() {
		delete(clients, c)
		log.Println("Closing Websocket")
		c.Close()
	}()
	clients[c] = struct{}{}

	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error in reading json message. Error : %v", err)
			return
		}

		broadcast(msg)
	}
}
