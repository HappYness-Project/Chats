package repository

import (
	"database/sql"
	"time"

	"github.com/HappYness-Project/chatApi/internal/chat/domain"
	"github.com/google/uuid"
)

type ChatRepository interface {
	GetChatByUserGroupId(userGroupId int) (*domain.Chat, error)
	GetChatById(chatId string) (*domain.Chat, error)
	GetChatByGroupID(groupID int) (*domain.Chat, error)
	CreateChat(chat *domain.Chat) (*domain.Chat, error)
	CreateChatWithParticipant(chat *domain.Chat, participant *domain.ChatParticipant) (*domain.Chat, error)
	DeleteChat(chatId string) error
	GetChatParticipants(chatId string) ([]domain.ChatParticipant, error)
	AddParticipantToChat(participant *domain.ChatParticipant) (*domain.ChatParticipant, error)
	IsUserParticipantInChat(chatId, userId string) (bool, error)
	DeleteParticipantFromChat(chatId, participantId string) error
}

type ChatRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) GetChatById(chatId string) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE id = $1`, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	chat := new(domain.Chat)
	for rows.Next() {
		chat, err = scanRowsIntoChat(rows)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepo) GetChatByUserGroupId(userGroupId int) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE usergroup_id = $1 and type = 'group'`, userGroupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chat := new(domain.Chat)
	for rows.Next() {
		chat, err = scanRowsIntoChat(rows)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepo) GetChatByGroupID(groupID int) (*domain.Chat, error) {
	rows, err := r.db.Query(`SELECT id, type, usergroup_id, container_id, created_at
							FROM public.chat
							WHERE usergroup_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chat := new(domain.Chat)
	for rows.Next() {
		chat, err = scanRowsIntoChat(rows)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepo) CreateChat(chat *domain.Chat) (*domain.Chat, error) {
	_, err := r.db.Exec(`INSERT INTO public.chat (id, type, usergroup_id, container_id, created_at)
						VALUES ($1, $2, $3, $4, $5)`,
		chat.Id, chat.Type.String(), chat.UserGroupId, chat.ContainerId, chat.CreatedAt)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (r *ChatRepo) CreateChatWithParticipant(chat *domain.Chat, participant *domain.ChatParticipant) (*domain.Chat, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create chat
	_, err = tx.Exec(`INSERT INTO public.chat (id, type, usergroup_id, container_id, created_at)
					  VALUES ($1, $2, $3, $4, $5)`,
		chat.Id, chat.Type.String(), chat.UserGroupId, chat.ContainerId, chat.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Generate participant ID and timestamp
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	participant.Id = id.String()
	participant.JoinedAt = time.Now().UTC()

	// Add participant
	_, err = tx.Exec(`INSERT INTO public.chat_participant (id, chat_id, user_id, joined_at, role, status)
					  VALUES ($1, $2, $3, $4, $5, $6)`,
		participant.Id, participant.ChatId, participant.UserId, participant.JoinedAt,
		participant.Role.String(), participant.Status.String())
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return chat, nil
}

func (r *ChatRepo) DeleteChat(chatId string) error {
	_, err := r.db.Exec(`DELETE FROM public.chat WHERE id = $1`, chatId)
	return err
}

func (r *ChatRepo) GetChatParticipants(chatId string) ([]domain.ChatParticipant, error) {
	rows, err := r.db.Query(`SELECT id, chat_id, user_id, joined_at, role, status
							FROM public.chat_participant
							WHERE chat_id = $1
							ORDER BY joined_at ASC`, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := make([]domain.ChatParticipant, 0)
	for rows.Next() {
		participant, err := scanRowsIntoChatParticipant(rows)
		if err != nil {
			return nil, err
		}
		participants = append(participants, *participant)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *ChatRepo) IsUserParticipantInChat(chatId, userId string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM public.chat_participant
						  WHERE chat_id = $1 AND user_id = $2`, chatId, userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ChatRepo) AddParticipantToChat(participant *domain.ChatParticipant) (*domain.ChatParticipant, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	participant.Id = id.String()
	participant.JoinedAt = time.Now().UTC()

	_, err = r.db.Exec(`INSERT INTO public.chat_participant (id, chat_id, user_id, joined_at, role, status)
						VALUES ($1, $2, $3, $4, $5, $6)`,
		participant.Id, participant.ChatId, participant.UserId, participant.JoinedAt,
		participant.Role.String(), participant.Status.String())
	if err != nil {
		return nil, err
	}

	return participant, nil
}

func (r *ChatRepo) DeleteParticipantFromChat(chatId, participantId string) error {
	_, err := r.db.Exec(`DELETE FROM public.chat_participant
						 WHERE chat_id = $1 AND user_id = $2`, chatId, participantId)
	return err
}

func scanRowsIntoChat(rows *sql.Rows) (*domain.Chat, error) {
	chat := new(domain.Chat)
	var typeStr string
	err := rows.Scan(
		&chat.Id,
		&typeStr,
		&chat.UserGroupId,
		&chat.ContainerId,
		&chat.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	chat.Type = domain.ChatType(typeStr)
	return chat, nil
}

func scanRowsIntoChatParticipant(rows *sql.Rows) (*domain.ChatParticipant, error) {
	var id, chatId, userId string
	var joinedAt time.Time
	var roleStr, statusStr string

	err := rows.Scan(&id, &chatId, &userId, &joinedAt, &roleStr, &statusStr)
	if err != nil {
		return nil, err
	}

	role, err := domain.NewParticipantRole(roleStr)
	if err != nil {
		return nil, err
	}

	status, err := domain.NewParticipantStatus(statusStr)
	if err != nil {
		return nil, err
	}

	return &domain.ChatParticipant{
		Id:       id,
		ChatId:   chatId,
		UserId:   userId,
		JoinedAt: joinedAt,
		Role:     role,
		Status:   status,
	}, nil
}
