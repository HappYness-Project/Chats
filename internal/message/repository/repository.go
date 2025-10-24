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
	SoftDelete(messageID, userID string) error
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
		SELECT id, chat_id, sender_id, content, message_type, created_at, read_status, deleted_at, deleted_by
		FROM (
			SELECT id, chat_id, sender_id, content, message_type, created_at, read_status, deleted_at, deleted_by
			FROM message
			WHERE chat_id = $1 AND deleted_at IS NULL
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
			&msg.MessageType, &msg.CreatedAt, &msg.ReadStatus, &msg.DeletedAt, &msg.DeletedBy)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *MessageRepo) SoftDelete(messageID, userID string) error {
	query := `
		UPDATE message
		SET deleted_at = CURRENT_TIMESTAMP, deleted_by = $2
		WHERE id = $1 AND sender_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, messageID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
