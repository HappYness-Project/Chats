package route

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	domain "github.com/HappYness-Project/chatApi/internal/message/domain"
	"github.com/HappYness-Project/chatApi/loggers"
	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	clients   map[*websocket.Conn]bool
	broadcast chan domain.Message
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex
	logger    *loggers.AppLogger
}

func NewWebSocketManager(logger *loggers.AppLogger) *WebSocketManager {
	return &WebSocketManager{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan domain.Message, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logger: logger,
	}
}

func (wsm *WebSocketManager) AddClient(conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	wsm.clients[conn] = true
	wsm.logger.Info().Msg("New client connected - clients: " + conn.LocalAddr().String())

}

func (wsm *WebSocketManager) RemoveClient(conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	delete(wsm.clients, conn)
	conn.Close()
	wsm.logger.Info().Msg("Client disconnected - client number: " + strconv.Itoa(len(wsm.clients)))
}

func (wsm *WebSocketManager) BroadcastMessage(msg domain.Message) {
	select {
	case wsm.broadcast <- msg:
	default:
		fmt.Println("Broadcast channel full, dropping message")
	}
}

func (wsm *WebSocketManager) SendToClients(msg domain.Message, logger *loggers.AppLogger) {
	wsm.mutex.RLock()
	clientsCopy := make([]*websocket.Conn, 0, len(wsm.clients))
	for client := range wsm.clients {
		clientsCopy = append(clientsCopy, client)
	}
	wsm.mutex.RUnlock()

	for _, client := range clientsCopy {
		err := client.WriteJSON(msg)
		if err != nil {
			logger.Error().Err(err).Msg("Unable to write a message")
			wsm.RemoveClient(client)
		}
	}
}
