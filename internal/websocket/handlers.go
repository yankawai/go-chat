package websocket

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// http -> websocketx
var Upgrader = websocket.Upgrader{} 

//  http converter to websocket!
func WSHandler(hub *Hub ,w http.ResponseWriter, r *http.Request) { 
	c, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[🚨] Upgrade error: ",err)
		return
	}

	log.Println("✅ Client Connected!")
	// creating user id 
	id := uuid.New().String()

	// add client in our mini-database [hub]
	hub.AddClient(c,id)

		// checking user connection: read/send "msg" [CHAT-logic]
		for { 
			_,msg,err := c.ReadMessage()
			if err != nil { 
				hub.DeleteClient(c)
				break
			}
			hub.BroadCast(msg)
		}
}