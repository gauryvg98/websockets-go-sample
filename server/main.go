package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var connections = struct {
	sync.RWMutex
	conns map[*websocket.Conn]bool
}{conns: make(map[*websocket.Conn]bool)}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}(conn)

	// Add the new connection to the map
	connections.Lock()
	connections.conns[conn] = true
	connections.Unlock()

	log.Println("Client Connected")

	// Remove the connection when the function exits
	defer func() {
		connections.Lock()
		delete(connections.conns, conn)
		connections.Unlock()
		log.Println("Client Disconnected")
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		broadcastMessage(messageType, message)
	}
}

func broadcastMessage(messageType int, message []byte) {
	connections.RLock()
	defer connections.RUnlock()
	for conn := range connections.conns {
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println(err)
			err := conn.Close()
			connections.Lock()
			delete(connections.conns, conn)
			connections.Unlock()
			if err != nil {
				return
			}
		}
		fmt.Println(string(message), messageType)
	}
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("WebSocket server started at :8080")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
