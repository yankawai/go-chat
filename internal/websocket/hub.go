package websocket

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Hub struct { 
	Clients  map[*websocket.Conn]string
}

// just return newhub (data base actually connecteds)
func NewHub() *Hub { 
	return  &Hub{
		Clients: make(map[*websocket.Conn]string),
	}
}

// user connected method [hub]
func (h *Hub) AddClient(c *websocket.Conn,id string) { 
	h.Clients[c] = id 
}

// user disconnected method [hub]
func (h *Hub) DeleteClient(c *websocket.Conn)  {
	log.Println("🔴 Client Disconnected!")
	delete(h.Clients,c)
	c.Close()
}

//
func (h *Hub) BroadCast(msg []byte)    { 
	for conn := range h.Clients.Clients { 
		if err := conn.WriteMessage(websocket.TextMessage,msg); err != nil { 
			fmt.Println("[⚠️] Sending Error: ",err)
		}
	}
}
