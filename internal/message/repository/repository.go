package repository

import (
	"database/sql"

	domain "github.com/HappYness-Project/chatApi/internal/message/domain"
)

type MessageRepository interface {
	Create(message domain.Message) error
	GetByChatID(chatID string, limit, offset int) ([]domain.Message, error)
	GetByGroupId(groupID int, limit, offset int) ([]domain.Message, error)
	GetByUserGroup(userIDs []string, limit, offset int) ([]domain.Message, error)
}

type MessageRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}
func (r *MessageRepo) Create(message domain.Message) error {
	query := `
		INSERT INTO message (id, chat_id, sender_id, content, message_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, message.ID, message.ChatID, message.SenderID, message.Content, message.MessageType, message.CreatedAt)

	return err
}
func (r *MessageRepo) GetByChatID(chatID string, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, message_type, created_at, read_status
		FROM (
			SELECT id, chat_id, sender_id, content, message_type, created_at, read_status
			FROM message
			WHERE chat_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		) AS recent_messages
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content,
			&msg.MessageType, &msg.CreatedAt, &msg.ReadStatus)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
