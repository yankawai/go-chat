package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func ws(w http.ResponseWriter, r *http.Request)  {
	c,err := upgrader.Upgrade(w,r,nil)
	if err != nil { 
		return
	}
	defer c.Close()
	for { 
		mt,msg,err  := c.ReadMessage()
		if err != nil {break}
		if  err = c.WriteMessage(mt,msg);err != nil{break}
	}
}

func main() {
	http.HandleFunc("/ws",ws)
	http.Handle("/",http.FileServer(http.Dir(".")))
	fmt.Println("server hosted on: http://localhost:8080/ws")
	err := http.ListenAndServe(":8080",nil) 	
	if err != nil { 
		panic(err)
	}
}