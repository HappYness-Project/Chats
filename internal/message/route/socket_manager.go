package route

import (
	"fmt"
	"net/http"
	"sync"

	domain "github.com/HappYness-Project/chatApi/internal/message/domain"
	"github.com/HappYness-Project/chatApi/loggers"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	clients   map[uuid.UUID][]*websocket.Conn
	broadcast chan domain.Message
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex
	logger    *loggers.AppLogger
}

func NewWebSocketManager(logger *loggers.AppLogger) *WebSocketManager {
	return &WebSocketManager{
		clients:   make(map[uuid.UUID][]*websocket.Conn),
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

func (wsm *WebSocketManager) AddClient(chatID uuid.UUID, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	wsm.clients[chatID] = append(wsm.clients[chatID], conn)
	wsm.logger.Info().Str("chatID", chatID.String()).Msg("New client connected")
}

func (wsm *WebSocketManager) RemoveClient(chatID uuid.UUID, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	conns := wsm.clients[chatID]
	for i, c := range conns {
		if c == conn {
			wsm.clients[chatID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(wsm.clients[chatID]) == 0 {
		delete(wsm.clients, chatID)
	}
	conn.Close()
	wsm.logger.Info().Str("chatID", chatID.String()).Msg("Client disconnected")
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
	conns := wsm.clients[msg.ChatID]
	clientsCopy := make([]*websocket.Conn, len(conns))
	copy(clientsCopy, conns)
	wsm.mutex.RUnlock()

	for _, client := range clientsCopy {
		err := client.WriteJSON(msg)
		if err != nil {
			logger.Error().Err(err).Msg("Unable to write a message")
			wsm.RemoveClient(msg.ChatID, client)
		}
	}
}
