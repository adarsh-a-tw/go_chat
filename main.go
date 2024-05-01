package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const channel = "chat"

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	go func() {
		ctx := context.Background()
		sub := rdb.Subscribe(ctx, channel)
		for {
			message, err := sub.ReceiveMessage(ctx)
			if err != nil {
				fmt.Println("Error receiving message", err)
			}
			if message != nil {
				broadcast([]byte(message.Payload))
			}
		}
	}()

	router := gin.Default()
	router.StaticFile("/", "./static/index.html")
	router.GET("/ws", serveWs(rdb))
	err := router.Run()
	if err != nil {
		log.Fatalf("Unable to start server. Error %v", err)
	}
	log.Println("Server started successfully.")
}

func serveWs(rdb *redis.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Error in upgrading web socket. Error: %v", err)
			return
		}

		go handleClient(conn, rdb)
	}
}

var clients = make(map[*websocket.Conn]struct{})

type Message struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

func broadcast(msgBytes []byte) {
	for conn := range clients {
		conn.WriteMessage(websocket.TextMessage, msgBytes)
	}
}

func handleClient(c *websocket.Conn, rdb *redis.Client) {
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

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Err marshaling", err.Error())
			return
		}

		err = rdb.Publish(context.Background(), channel, string(msgBytes)).Err()
		if err != nil {
			fmt.Println("Error publishing:", err.Error())
		}
	}
}
