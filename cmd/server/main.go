package main

import (
	ws "go-chat/internal/websocket"
	"log"

	"github.com/gin-gonic/gin"
)


func main()  {
	// init [tools for host & control server]
	router := gin.New()
	hub := ws.NewHub()

	// http [download index.html] - #UserPage
	router.GET("/",func(c *gin.Context) {
		c.File("./static/index.html")
	})
	router.Static("/static","./static")
	// websocket 
	router.GET("/ws", func(c *gin.Context) {
		ws.WSHandler(hub,c.Writer,c.Request)
	})

	log.Println("🟢 Server Online: http://localhost:8080")
	if err := router.Run(":8080");err != nil { 
		log.Fatal("[🔴] Failed to start server: ", err)
	}
}
