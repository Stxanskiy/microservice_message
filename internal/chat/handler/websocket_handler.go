package handler

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"gitlab.com/nevasik7/lg"
	"serice_message/internal/chat/model"
	"serice_message/internal/chat/uc"
	"serice_message/pkg/jwt"

	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	JWTManager *jwt.JWTManager
	MessageUC  *uc.MessageUC
}

func NewWebSocketHandler(jwtManager *jwt.JWTManager, messageUC *uc.MessageUC) *WebSocketHandler {
	return &WebSocketHandler{
		JWTManager: jwtManager,
		MessageUC:  messageUC,
	}
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	//Поднимаем WebSocket-соединение
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Auth Через JWT
	token := r.URL.Query().Get("token")
	claims, err := h.JWTManager.VerifyToken(token)
	if err != nil {
		lg.Errorf("Unauthorized WebSocket connection attempt:%v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		return
	}

	userID := claims["user_id"].(float64)
	// TODO Забыть убрать INFOF и поставить обычный Printf
	lg.Infof("User %d connected to WebSocket\n", int(userID))

	//Обработка сообщений
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			lg.Errorf("Error reading meesage: %v\n", err)
			break
		}
		// TODO решить какой уровень логирования указать для данного сообщение
		lg.Printf("Reciver message from user %d %s\n", int(userID), string(message))

		//Эхо ответ для тестирования работоспособности
		//TODO не забыть убрать эхо ответ после тестирования
		if err := conn.WriteMessage(messageType, message); err != nil {
			lg.Errorf("Error sending message:%v\n", err)
			break
		}

		var msg model.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			lg.Errorf("Invalid message fromat %v", err, string(message))
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid message format"))
			continue
		}

		//Валидация обязателых полей
		if msg.ChatID == 0 || msg.Content == "" {
			lg.Errorf("Missing required fields: chat_id=%d, content=%s", msg.ChatID, msg.Content)
			conn.WriteMessage(websocket.TextMessage, []byte("Missing required fields"))
			continue
		}

		//Устновка дополнительныйх полей
		msg.SenderId = int(userID)
		msg.Timestamp = time.Now().Format(time.RFC3339)

		//передача уведомления в kafka
		if err := h.MessageUC.HandleNewMessage(context.Background(), msg); err != nil {
			lg.Errorf("Failed to handle message:%v", err)
			continue
		}

		//TODO не забыть поменять уровень логирования
		lg.Infof("Message from user %d processed: %s", userID, msg.Content)
	}

}
