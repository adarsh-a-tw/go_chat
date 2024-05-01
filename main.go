package main

import (
	"io"
	"log"
	"net/http"

	gripcontrol "github.com/fanout/go-gripcontrol"
	"github.com/fanout/go-pubcontrol"
	"github.com/gin-gonic/gin"
)

var pub = gripcontrol.NewGripPubControl([]map[string]interface{}{
	{"control_uri": "http://localhost:5561"},
})

func main() {

	router := gin.Default()
	router.StaticFile("/", "./static/index.html")
	router.POST("/ws", serveWs)
	err := router.Run()
	if err != nil {
		log.Fatalf("Unable to start server. Error %v", err)
	}
	log.Println("Server started successfully.")
}

func serveWs(c *gin.Context) {
	c.Header("Sec-WebSocket-Extensions", "grip")
	c.Header("Content-Type", "application/websocket-events")
	// c.String(http.StatusOK, "OPEN\r\n")
	body, _ := io.ReadAll(c.Request.Body)
	inEvents, err := gripcontrol.DecodeWebSocketEvents(string(body))
	if err != nil {
		panic("Failed to decode WebSocket events: " + err.Error())
	}

	if inEvents[0].Type == "OPEN" {
		wsControlMessage, err := gripcontrol.WebSocketControlMessage("subscribe",
			map[string]interface{}{"channel": "chat"})
		if err != nil {
			panic("Unable to create control message: " + err.Error())
		}

		// Open the WebSocket and subscribe it to a channel:
		outEvents := []*gripcontrol.WebSocketEvent{
			{Type: "OPEN"},
			{Type: "TEXT",
				Content: "c:" + wsControlMessage}}
		c.String(http.StatusOK, gripcontrol.EncodeWebSocketEvents(outEvents))
		return
	}

	if inEvents[0].Type == "TEXT" {
		format := &gripcontrol.WebSocketMessageFormat{
			Content: []byte(inEvents[0].Content)}
		item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")
		err = pub.Publish("chat", item)
		if err != nil {
			panic("Publish failed with: " + err.Error())
		}
	}
}