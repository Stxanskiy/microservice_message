package route

import (
	"github.com/go-chi/chi/v5"
	"serice_message/internal/chat/handler"
)

func RegisterWebSocketRoutes(r chi.Router, handler *handler.WebSocketHandler) {
	r.HandleFunc("/ws", handler.HandleConnection)
}
