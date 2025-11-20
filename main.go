package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)


var upgrader = websocket.Upgrader{}            // websocket
var clients = make(map[*websocket.Conn]string) // global client structure

func ws(w http.ResponseWriter, r *http.Request) { //  http converter to websocket!
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	id := uuid.New().String()
	clients[c] = id

	log.Println("✅ Client Connected!")
	defer func() {
		delete(clients, c)
		log.Println("🔌 Client Disconnected!")
		c.Close()
	}()

	// READING/CHECK/SEND USER MESSAGE ->
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// CHAT [BROADCAST]
		for conn := range clients {
			if err := conn.WriteMessage(mt, []byte(string(msg))); err != nil {
				log.Println("Ошибка отправки!")
			}
		}
	}
}


func main() {
	router  := gin.Default() 

	// Handler for HTML (frontend)
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	router.Static("/static","./static")

	// Websocket 
	router.GET("/ws",func(c *gin.Context) {
		ws(c.Writer,c.Request) // http -> websocket [func]
})

	// Host srver
	fmt.Println("🏠 Server hosted on: http://localhost:8080/")
	if err := router.Run(":8080"); err != nil  {
		log.Fatal("Failed to start server:", err)
	}
}
