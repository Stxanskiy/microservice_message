package repo

const (
	QueryCreateChat        = `INSERT INTO chats (name, type, created_at) VALUES ($1, $2, $3) RETURNING id`
	QueryAddParticipant    = `INSERT INTO chat_participants (chat_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)`
	QueryRemoveParticipant = `DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2`
	QuerySaveMessage       = `INSERT INTO messages (chat_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	QueryGetChatHistory    = `SELECT id, chat_id, sender_id, content, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at ASC`
	QueryCachedMessage     = `INSERT INTO message_cache (chat_id, message_id, sender_id, content, created_at) VALUES (?, ?, ?, ?, ?)`
	QueryGetCachedMessages = `SELECT message_id, chat_id, sender_id, content, created_at FROM message_cache WHERE chat_id = ?`
	QueryGetChatsUserID    = ` SELECT c.id, c.name, c.type, c.created_at FROM chats c JOIN chat_participants cp ON c.id = cp.chat_id WHERE cp.user_id = $1`
	QueryGetNicknameUserID = `SELECT u.nickname FROM chat_participants cp INNER JOIN users u ON u.id = cp.user_id WHERE cp.chat_id = $1 AND cp.user_id <> $2 LIMIT 1`
	QueryGetPrivateChat    = `SELECT c.id, c.name, c.type, c.created_at FROM chats c INNER JOIN chat_participants cp1 ON cp1.chat_id = c.id INNER JOIN chat_participants cp2 ON cp2.chat_id = c.id WHERE c.type = 'private' AND cp1.user_id = $1 AND cp2.user_id = $2`
	QueryGetChatsForUser   = `SELECT c.id, c.name, c.type FROM chats c INNER JOIN chat_participants cp ON cp.chat_id = c.id WHERE cp.user_id = $1`
)
