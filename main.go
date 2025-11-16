package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // websocket 
var clients = make(map[*websocket.Conn]string)  // global client structure 


func ws(w http.ResponseWriter, r *http.Request)  { //  http converter to websocket!
	c,err := upgrader.Upgrade(w,r,nil)
	if err != nil { 
		return
	}
	id :=   uuid.New().String()
	clients[c] =  id 	

	log.Println("🟩 Подключен Новый пользователь: ",id)
	defer func() {
		delete(clients,c)	
		log.Println("🔌 Пользователь отключился: ",id)
		c.Close()
	}()

	// READING/CHECK/SEND USER MESSAGE ->
	for { 
		mt,msg,err  := c.ReadMessage()
		if err != nil {
			break
		}

		// CHAT [BROADCAST]
		senderID :=  clients[c] 
		for conn := range clients { 
			if err := conn.WriteMessage(mt,[]byte(senderID+"-User: "+string(msg))); err != nil { 
				log.Println("Ошибка отправки!")
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws",ws)
	http.Handle("/",http.FileServer(http.Dir("./static")))
	fmt.Println("server hosted on: http://localhost:8080/ws")
	err := http.ListenAndServe(":8080",nil) 	
	if err != nil { 
		log.Println(err)
		return
	}
}

