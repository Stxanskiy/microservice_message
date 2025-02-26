package repo

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/nevasik7/lg"
	"sevice_message_1/intenral/chat/model"
	"time"
)

type ChatRepo struct {
	pgPool           *pgxpool.Pool
	cassandraSession *gocql.Session
}

func NewChatRepo(pgPool *pgxpool.Pool, cassandrasession *gocql.Session) *ChatRepo {
	return &ChatRepo{
		pgPool:           pgPool,
		cassandraSession: cassandrasession,
	}
}

// Todo Исправить потом в модел
type ChatDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	OtherNickname string `json:"otherNickname,omitempty"`
}

// create chat создание нового чата
func (r *ChatRepo) CreateChat(ctx context.Context, chat *model.Chat) error {
	query := QueryCreateChat
	err := r.pgPool.QueryRow(ctx, query, chat.Name, chat.Type, time.Now()).Scan(&chat.ID)
	if err != nil {
		lg.Errorf("Failed to create chat:%v", err)
	}
	return nil
}

// Добавление новых участников
func (r *ChatRepo) AddParticipant(ctx context.Context, participant *model.Participant) error {
	query := QueryAddParticipant
	_, err := r.pgPool.Exec(ctx, query, participant.ChatID, participant.UserID, participant.Role, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}
	return nil
}

// Удаление участников
func (r *ChatRepo) RemoveParticipant(ctx context.Context, chatID, userID int64) error {
	query := QueryRemoveParticipant
	_, err := r.pgPool.Exec(ctx, query, chatID, userID)
	if err != nil {
		lg.Errorf("Failed to remove participant:%v", err)
	}
	return nil
}

// SaveMessage удвляет участника из чата
func (r *ChatRepo) SaveMessage(ctx context.Context, msg *model.Message) error {
	query := QuerySaveMessage
	err := r.pgPool.QueryRow(ctx, query, msg.ChatID, msg.SenderID, msg.Content, time.Now()).Scan(&msg.ID)
	if err != nil {
		lg.Errorf("Faild to save message:%v", err)
	}
	// Асинхронно кэшируем сообщение в Cassandra
	go r.cacheMessage(msg)
	return nil
}

// GetChatHistory получает сообщения из чата
func (r *ChatRepo) GetChatHistory(ctx context.Context, chatID int64) ([]*model.Message, error) {
	// пробуем получить кэширлваие сообщения из Cassandra
	message, err := r.getCachedMessages(chatID)
	if err != nil && len(message) > 0 {
		return message, nil
	}

	// При отсутсвии кэша читаем из Postgres
	query := QueryGetChatHistory
	rows, err := r.pgPool.Query(ctx, query, chatID)
	if err != nil {
		lg.Errorf("Failed to get chat historu %v", err)
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.CreatedAt); err != nil {
			lg.Errorf("Failed to scan message:%w", err)
		}
		msgs = append(msgs, &msg)
	}
	//Ассинхронно кэшируем сообщения
	go r.cacheMessages(chatID, msgs)
	return msgs, nil
}

// получение списка чатов
// GetChatsForUser возвращает список чатов, в которых участвует пользователь с заданным userID.
func (r *ChatRepo) GetChatsForUser(ctx context.Context, userID int64) ([]ChatDTO, error) {
	// 1. Получаем все чаты пользователя
	query := `
		SELECT c.id, c.name, c.type
		FROM chats c
		INNER JOIN chat_participants cp ON cp.chat_id = c.id
		WHERE cp.user_id = $1
	`
	rows, err := r.pgPool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	var result []ChatDTO

	for rows.Next() {
		var chat ChatDTO
		if err := rows.Scan(&chat.ID, &chat.Name, &chat.Type); err != nil {
			return nil, err
		}

		// 2. Если чат приватный, находим второго участника
		if chat.Type == "private" {
			otherNickname, err := r.getOtherParticipantNickname(ctx, chat.ID, userID)
			if err != nil {
				return nil, err
			}
			chat.OtherNickname = otherNickname
		}

		result = append(result, chat)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

// getOtherParticipantNickname ищет никнейм второго участника приватного чата
func (r *ChatRepo) getOtherParticipantNickname(ctx context.Context, chatID, currentUserID int64) (string, error) {
	query := QueryGetNicknameUserID
	var nickname string
	err := r.pgPool.QueryRow(ctx, query, chatID, currentUserID).Scan(&nickname)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Нет второго участника — теоретически странная ситуация,
			// но вернём пустую строку
			return "", nil
		}
		return "", err
	}
	return nickname, nil
}

// repository/postgres_repo.go
func (r *ChatRepo) GetPrivateChatBetween(ctx context.Context, user1, user2 int64) (*model.Chat, error) {
	query := `
        SELECT c.id, c.name, c.type, c.created_at
        FROM chats c
        INNER JOIN chat_participants cp1 ON cp1.chat_id = c.id
        INNER JOIN chat_participants cp2 ON cp2.chat_id = c.id
        WHERE c.type = 'private'
          AND cp1.user_id = $1
          AND cp2.user_id = $2
    `
	row := r.pgPool.QueryRow(ctx, query, user1, user2)
	var chat model.Chat
	err := row.Scan(&chat.ID, &chat.Name, &chat.Type, &chat.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // нет существующего чата
		}
		return nil, err
	}
	return &chat, nil
}

// cacheMessage сохраняет одно сообщение в Cassandra (пример).
func (r *ChatRepo) cacheMessage(msg *model.Message) {
	query := QueryCachedMessage
	_ = r.cassandraSession.Query(query, msg.ChatID, msg.ID, msg.SenderID, msg.Content, msg.CreatedAt).Exec()
}

// getCachedMessages получает сообщения из Cassandra.
func (r *ChatRepo) getCachedMessages(chatID int64) ([]*model.Message, error) {
	query := QueryGetCachedMessages
	iter := r.cassandraSession.Query(query, chatID).Iter()
	var msgs []*model.Message
	var id, cID, senderID int64
	var content string
	var createdAt time.Time
	for iter.Scan(&id, &cID, &senderID, &content, &createdAt) {
		msgs = append(msgs, &model.Message{
			ID:        id,
			ChatID:    cID,
			SenderID:  senderID,
			Content:   content,
			CreatedAt: createdAt,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return msgs, nil
}

// cacheMessages кэширует пакет сообщений.
func (r *ChatRepo) cacheMessages(chatID int64, messages []*model.Message) {
	for _, msg := range messages {
		r.cacheMessage(msg)
	}
}
