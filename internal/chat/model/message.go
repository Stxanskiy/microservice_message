package model

type Message struct {
	ChatID    int    `json:"chat_id"`
	SenderId  int    `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}
