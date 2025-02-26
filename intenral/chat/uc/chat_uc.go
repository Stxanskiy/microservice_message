package uc

import (
	"context"
	"fmt"
	"gitlab.com/nevasik7/lg"
	"sevice_message_1/intenral/chat/model"
	"sevice_message_1/intenral/chat/repo"
)

type ChatUCIn interface {
	CreatePrivateChat(ctx context.Context, user1, user2 int64) (*model.Chat, error)
	CreateGroupChat(ctx context.Context, name string, adminID int64, participants []int64) (*model.Chat, error)
	AddParticipant(ctx context.Context, chatID, userID int64, role string) error
	RemoveParticipant(ctx context.Context, chatID, userID int64) error
	SendMessage(ctx context.Context, chatID, senderID int64, content string) (*model.Message, error)
	GetChatHistory(ctx context.Context, chatID int64) ([]*model.Message, error)
	GetChatsForUser(ctx context.Context, userID int64) ([]repo.ChatDTO, error)
}

type ChatUC struct {
	repo          *repo.ChatRepo
	kafkaProducer KafkaProducer
}

// KafkaProducer — интерфейс для публикации сообщений в Kafka.
type KafkaProducer interface {
	Publish(topic string, message []byte) error
}

func NewChatUC(repo *repo.ChatRepo, producer KafkaProducer) *ChatUC {
	return &ChatUC{
		repo:          repo,
		kafkaProducer: producer,
	}
}

// CreatePrivateChat создаёт приватный чат между двумя пользователями.
// Имя чата формируется автоматически, например "chat_1_2".
func (uc *ChatUC) CreatePrivateChat(ctx context.Context, user1, user2 int64) (*model.Chat, error) {
	// 1. Проверяем, существует ли уже приватный чат
	existingChat, err := uc.repo.GetPrivateChatBetween(ctx, user1, user2)
	if err != nil {
		return nil, err
	}
	if existingChat != nil {
		// уже существует
		return existingChat, nil
	}
	//Если нет такого чата создаем
	name := fmt.Sprintf("chat_%d_%d", user1, user2)
	chat := &model.Chat{
		Name: name,
		Type: model.PrivateChat,
	}
	if err := uc.repo.CreateChat(ctx, chat); err != nil {
		lg.Errorf("Failed to create private chat")
	}
	// Добавляем обоих участников
	if err := uc.repo.AddParticipant(ctx, &model.Participant{ChatID: chat.ID, UserID: user1, Role: "member"}); err != nil {
		lg.Errorf("Failed to paticipants private chat %v,", err)
	}
	if err := uc.repo.AddParticipant(ctx, &model.Participant{ChatID: chat.ID, UserID: user2, Role: "member"}); err != nil {
		lg.Errorf("Failed add paticipants private chat %v", err)
	}
	return chat, nil
}

// CreateGroupChat создаёт групповой чат с заданным названием.
func (uc *ChatUC) CreateGroupChat(ctx context.Context, name string, adminID int64, participants []int64) (*model.Chat, error) {
	chat := &model.Chat{
		Name: name,
		Type: model.GroupChat,
	}
	if err := uc.repo.CreateChat(ctx, chat); err != nil {
		lg.Errorf("Failed create group chat %v", err)
	}
	// Добавляем администратора
	if err := uc.repo.AddParticipant(ctx, &model.Participant{ChatID: chat.ID, UserID: adminID, Role: "admin"}); err != nil {
		lg.Errorf("Failed to add Admin group chat %v", err)
	}
	// Добавляем остальных участников
	for _, userID := range participants {
		if err := uc.repo.AddParticipant(ctx, &model.Participant{ChatID: chat.ID, UserID: userID, Role: "member"}); err != nil {
			lg.Errorf("Failed to add participants group chat %v", err)
		}
	}
	return chat, nil
}

func (uc *ChatUC) AddParticipant(ctx context.Context, chatID, userID int64, role string) error {
	return uc.repo.AddParticipant(ctx, &model.Participant{ChatID: chatID, UserID: userID, Role: role})
}

func (uc *ChatUC) RemoveParticipant(ctx context.Context, chatID, userID int64) error {
	return uc.repo.RemoveParticipant(ctx, chatID, userID)
}

func (uc *ChatUC) SendMessage(ctx context.Context, chatID, senderID int64, content string) (*model.Message, error) {
	if content == "" {
		lg.Errorf("message content cannot be empty")
	}
	msg := &model.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
	}
	if err := uc.repo.SaveMessage(ctx, msg); err != nil {
		lg.Errorf("Failed Save Message:%v", err)
	}
	// Публикуем событие в Kafka (например, для уведомлений)
	kafkaMsg := []byte(fmt.Sprintf("ChatID:%d,SenderID:%d,Message:%s", chatID, senderID, content))
	if err := uc.kafkaProducer.Publish("messages", kafkaMsg); err != nil {
		lg.Errorf("Failed publish Message %v", err)
	}
	return msg, nil
}

func (uc *ChatUC) GetChatHistory(ctx context.Context, chatID int64) ([]*model.Message, error) {
	return uc.repo.GetChatHistory(ctx, chatID)
}

func (uc *ChatUC) GetChatsForUser(ctx context.Context, userID int64) ([]repo.ChatDTO, error) {
	return uc.repo.GetChatsForUser(ctx, userID)
}
