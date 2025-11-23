package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct { 
	Conn *websocket.Conn
	ID string
	Hub *Hub
}


func (c* Client) Read() {
	defer c.Hub.DeleteClient(c.Conn)

	for { 
		_,msg,err := c.Conn.ReadMessage()
		if err != nil { 
			break	
		}
		c.Hub.BroadCast(msg)
	}
}

func (c *Client) Send(msg []byte) { 
	if err := c.Conn.WriteMessage(websocket.TextMessage,msg); err != nil {
		log.Println("[⚠️] Send Error:", err)
		c.Hub.DeleteClient(c.Conn)
	}
}
