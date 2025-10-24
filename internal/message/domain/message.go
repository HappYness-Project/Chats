package entity

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID  `json:"id"`
	ChatID      uuid.UUID  `json:"chat_id"`
	SenderID    uuid.UUID  `json:"sender_id"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	CreatedAt   time.Time  `json:"created_at"`
	ReadStatus  bool       `json:"read_status"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	DeletedBy   *uuid.UUID `json:"deleted_by,omitempty"`
}
