package http

import (
	"encoding/json"
	"net/http"
	"sevice_message_1/intenral/chat/uc"
	"sevice_message_1/pkg/jwt"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ChatHandler struct {
	chatUC uc.ChatUCIn //Используем интерфейс а не конкретный тип
}

func NewChatHandler(u uc.ChatUCIn) *ChatHandler {
	return &ChatHandler{chatUC: u}
}

// NewRouter возвращает роутер с зарегистрированными эндпоинтами для работы с чатами.
func NewRouter(u uc.ChatUCIn) http.Handler {
	r := chi.NewRouter()
	handler := NewChatHandler(u)
	r.Post("/private", handler.CreatePrivateChat)
	r.Post("/group", handler.CreateGroupChat)
	r.Post("/{chatID}/message", handler.SendMessage)
	r.Get("/{chatID}/history", handler.GetChatHistory)
	// Дополнительные эндпоинты для управления участниками
	r.Post("/{chatID}/add", handler.AddParticipant)
	r.Post("/{chatID}/remove", handler.RemoveParticipant)
	r.Get("/get_chats", handler.GetChats)
	return r
}

// CreatePrivateChatRequest для создания приватного чата.
type CreatePrivateChatRequest struct {
	User1 int64 `json:"user1"`
	User2 int64 `json:"user2"`
}

func (h *ChatHandler) CreatePrivateChat(w http.ResponseWriter, r *http.Request) {
	var req CreatePrivateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	chat, err := h.chatUC.CreatePrivateChat(r.Context(), req.User1, req.User2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chat)
}

// CreateGroupChatRequest для создания группового чата.
type CreateGroupChatRequest struct {
	Name         string  `json:"name"`
	AdminID      int64   `json:"admin_id"`
	Participants []int64 `json:"participants"`
}

func (h *ChatHandler) CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	chat, err := h.chatUC.CreateGroupChat(r.Context(), req.Name, req.AdminID, req.Participants)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chat)
}

// SendMessageRequest для отправки сообщения.
type SendMessageRequest struct {
	SenderID int64  `json:"sender_id"`
	Content  string `json:"content"`
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	msg, err := h.chatUC.SendMessage(r.Context(), chatID, req.SenderID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(msg)
}

func (h *ChatHandler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	messages, err := h.chatUC.GetChatHistory(r.Context(), chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}

// AddParticipantRequest для добавления участника.
type AddParticipantRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"` // например, "member" или "admin"
}

func (h *ChatHandler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	var req AddParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.chatUC.AddParticipant(r.Context(), chatID, req.UserID, req.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Participant added"})
}

func (h *ChatHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	var req AddParticipantRequest // тот же формат, что и для добавления
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.chatUC.RemoveParticipant(r.Context(), chatID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Participant removed"})
}

// Новый endpoint: GetChats возвращает список чатов для аутентифицированного пользователя.
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chats, err := h.chatUC.GetChatsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

// Пример: извлекаем userID из контекста
func getUserIDFromContext(r *http.Request) (int64, bool) {
	val := r.Context().Value(jwt.ContextUserIDKey)
	if val == nil {
		return 0, false
	}
	userID, ok := val.(int64)
	return userID, ok
}
