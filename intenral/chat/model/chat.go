package model

import "time"

// Типы чатов: private — для двух пользователей, group — для группового чата.
type ChatType string

const (
	PrivateChat ChatType = "private"
	GroupChat   ChatType = "group"
)

// Chat — сущность чата.
type Chat struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      ChatType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// Participant — участник чата.
type Participant struct {
	ChatID   int64     `json:"chat_id"`
	UserID   int64     `json:"user_id"`
	Role     string    `json:"role"` // "admin" или "member"
	JoinedAt time.Time `json:"joined_at"`
}

// Message — сообщение в чате.
type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
