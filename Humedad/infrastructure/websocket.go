// websocket.go
package infrastructure

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("New WebSocket client connected. Total clients: %d", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}
		case message := <-h.broadcast:
			log.Printf("Broadcasting message to %d clients: %s", len(h.clients), string(message))
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	h.register <- conn

	defer func() {
		h.unregister <- conn
	}()

	// Mantener la conexión abierta
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
