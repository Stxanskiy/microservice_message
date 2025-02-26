package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"gitlab.com/nevasik7/lg"

	"sevice_message_1/intenral/chat/model"
	"sevice_message_1/intenral/chat/uc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	chatUC uc.ChatUCIn
	hub    *Hub
}

func NewWebSocketHandler(u uc.ChatUCIn, hub *Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chatIDStr := chi.URLParam(r, "chatID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Chat id", http.StatusBadRequest)
			lg.Errorf("Не удалось подключиться к чату. Err: %v", err)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			lg.Errorf("Ошибка при апгрейде WebSocket: %v", err)
			return
		}
		handleConnection(r.Context(), conn, chatID, u, hub)
	})
}

func handleConnection(ctx context.Context, conn *websocket.Conn, chatID int64, chatUC uc.ChatUCIn, hub *Hub) {
	// 1. Регистрируем сокет в хабе
	hub.Register <- subscription{
		ChatID: chatID,
		Conn:   conn,
	}

	defer func() {
		// При выходе/ошибке — отписываем
		hub.Unregister <- subscription{
			ChatID: chatID,
			Conn:   conn,
		}
	}()

	// 2. Читаем сообщения в цикле
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			lg.Errorf("WebSocket read error: %v", err)
			break
		}

		// Если хотите, можно сначала распарсить JSON, сохранить в БД и т.д.
		var msg model.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			lg.Errorf("Invalid message format: %v ", err)
			continue
		}

		// Пример: сохраняем сообщение через use case
		savedMsg, saveErr := chatUC.SendMessage(ctx, chatID, msg.SenderID, msg.Content)
		if saveErr != nil {
			lg.Errorf("Ошибка сохранения сообщения: %v", saveErr)
			continue
		}

		// Преобразуем savedMsg в JSON для рассылки
		broadcastData, err := json.Marshal(savedMsg)
		if err != nil {
			lg.Errorf("Ошибка сериализации сообщения: %v", err)
			continue
		}

		// 3. Передаём сообщение всем сокетам в этой комнате (chatID)
		hub.Broadcast <- broadcastMessage{
			ChatID: chatID,
			Data:   broadcastData,
		}
	}
}
