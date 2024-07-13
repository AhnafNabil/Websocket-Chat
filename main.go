package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	clients[client] = true

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, client)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}