package websocket

import (
	"github.com/gorilla/websocket"
	"sync"
)

// HUB Управляет набором, сгруппированных по chatID
type Hub struct {
	// Для каждой комнаты (chatID) храним набор подключений
	rooms map[int64]map[*websocket.Conn]bool

	// Каналы для управления подключениями
	Register   chan subscription
	Unregister chan subscription

	// Канал для широковещательной рассылки (в рамках одной комнаты)
	Broadcast chan broadcastMessage

	mu sync.RWMutex
}

// subscription описывает запрос на регистрацию/отключение сокета к определённой комнате
type subscription struct {
	ChatID int64
	Conn   *websocket.Conn
}

type broadcastMessage struct {
	ChatID int64
	Data   []byte
}

// NewHub инициализирует новый хаб
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*websocket.Conn]bool),
		Register:   make(chan subscription),
		Unregister: make(chan subscription),
		Broadcast:  make(chan broadcastMessage),
	}
}

// Run запускает бесконечный цикл обработки событий (регистраций, отключений, рассылок).
func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.Register:
			h.registerConnection(sub)
		case sub := <-h.Unregister:
			h.unregisterConnection(sub)
		case msg := <-h.Broadcast:
			h.broadcastToRoom(msg)
		}
	}
}

// registerConnection добавляет соединение в нужную комнату.
func (h *Hub) registerConnection(sub subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[sub.ChatID] == nil {
		h.rooms[sub.ChatID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[sub.ChatID][sub.Conn] = true
}

// unregisterConnection удаляет соединение из комнаты.
func (h *Hub) unregisterConnection(sub subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.rooms[sub.ChatID]
	if conns != nil {
		if _, ok := conns[sub.Conn]; ok {
			delete(conns, sub.Conn)
			if len(conns) == 0 {
				delete(h.rooms, sub.ChatID)
			}
		}
	}
	sub.Conn.Close()
}

// broadcastToRoom рассылает msg.Data всем подключениям в заданной комнате.
func (h *Hub) broadcastToRoom(msg broadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.rooms[msg.ChatID]
	for conn := range conns {
		err := conn.WriteMessage(websocket.TextMessage, msg.Data)
		if err != nil {
			// Если запись не удалась, убираем conn
			conn.Close()
			delete(conns, conn)
		}
	}
}
