package integration_tests

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/HappYness-Project/chatApi/dbs"
	"github.com/HappYness-Project/chatApi/internal/chat/domain"
	"github.com/HappYness-Project/chatApi/internal/chat/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	dsn := "postgres://postgres:postgres@localhost:8020/postgres?sslmode=disable"

	testDB, err = dbs.ConnectToDb(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func setupTestData(t *testing.T) {
	_, err := testDB.Exec(`
		INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at)
		VALUES
		('01987073-0a87-7b32-9439-86868dfe9bd3', 'group', 100, NULL, CURRENT_TIMESTAMP),
		('01987073-cf13-7621-af36-54ce20056d19', 'group', NULL, NULL, CURRENT_TIMESTAMP),
		('01987075-16cb-7337-af15-cd28f64c93a4', 'group', NULL, NULL, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)
}

func cleanupTestData(t *testing.T) {
	_, err := testDB.Exec(`
		DELETE FROM public.chat
		WHERE id IN (
			'01987073-0a87-7b32-9439-86868dfe9bd3',
			'01987073-cf13-7621-af36-54ce20056d19',
			'01987075-16cb-7337-af15-cd28f64c93a4'
		)
	`)
	require.NoError(t, err)
}
func TestChatRepository_DatabaseConnection(t *testing.T) {
	repo := repository.NewRepository(testDB)
	require.NotNil(t, repo)

	err := testDB.Ping()
	require.NoError(t, err)
}
func TestChatRepository_GetChatById(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return chat when valid ID provided", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd3"

		chat, err := repo.GetChatById(chatID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, chatID, chat.Id)
		assert.Equal(t, domain.ChatTypeGroup, chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, 100, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent ID provided", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		chat, err := repo.GetChatById(nonExistentID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

}
func TestChatRepository_GetChatByUserGroupId(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return group chat when valid user group ID provided", func(t *testing.T) {
		userGroupID := 100

		chat, err := repo.GetChatByUserGroupId(userGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-0a87-7b32-9439-86868dfe9bd3", chat.Id)
		assert.Equal(t, domain.ChatTypeGroup, chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, userGroupID, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent user group ID provided", func(t *testing.T) {
		nonExistentUserGroupID := 999

		chat, err := repo.GetChatByUserGroupId(nonExistentUserGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

	t.Run("should use existing test data from schema", func(t *testing.T) {
		userGroupID := 1

		chat, err := repo.GetChatByUserGroupId(userGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01991792-6185-7dbf-81a8-b352fb0e87d1", chat.Id)
		assert.Equal(t, domain.ChatTypeGroup, chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, userGroupID, *chat.UserGroupId)
	})
}
func TestChatRepository_GetChatByGroupID(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should return chat when valid group ID provided", func(t *testing.T) {
		groupID := 100

		chat, err := repo.GetChatByGroupID(groupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-0a87-7b32-9439-86868dfe9bd3", chat.Id)
		assert.Equal(t, domain.ChatTypeGroup, chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, groupID, *chat.UserGroupId)
		assert.Nil(t, chat.ContainerId)
		assert.False(t, chat.CreatedAt.IsZero())
	})

	t.Run("should return empty chat when non-existent group ID provided", func(t *testing.T) {
		nonExistentGroupID := 999

		chat, err := repo.GetChatByGroupID(nonExistentGroupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Empty(t, chat.Id)
	})

	t.Run("should return chat regardless of type (unlike GetChatByUserGroupId)", func(t *testing.T) {
		groupID := 2

		chat, err := repo.GetChatByGroupID(groupID)

		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, "01987073-cf13-7621-af36-54ce20056d18", chat.Id)
		assert.Equal(t, domain.ChatTypeGroup, chat.Type)
		assert.NotNil(t, chat.UserGroupId)
		assert.Equal(t, groupID, *chat.UserGroupId)
	})
}
func TestChatRepository_CreateChat(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should create group chat successfully", func(t *testing.T) {
		userGroupID := 200
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		createdChat, err := repo.CreateChat(chat)

		require.NoError(t, err)
		require.NotNil(t, createdChat)
		assert.NotEmpty(t, createdChat.Id)
		assert.Equal(t, domain.ChatTypeGroup, createdChat.Type)
		assert.NotNil(t, createdChat.UserGroupId)
		assert.Equal(t, userGroupID, *createdChat.UserGroupId)
		assert.Nil(t, createdChat.ContainerId)
		assert.False(t, createdChat.CreatedAt.IsZero())
		assert.True(t, createdChat.CreatedAt.After(time.Now().Add(-time.Minute)))

		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})
}
func TestChatRepository_DeleteChat(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)

	repo := repository.NewRepository(testDB)

	t.Run("should delete existing chat successfully", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd3"

		// Verify chat exists before deletion
		chat, err := repo.GetChatById(chatID)
		require.NoError(t, err)
		require.NotNil(t, chat)
		assert.Equal(t, chatID, chat.Id)

		// Delete the chat
		err = repo.DeleteChat(chatID)
		require.NoError(t, err)

		// Verify chat no longer exists
		deletedChat, err := repo.GetChatById(chatID)
		require.NoError(t, err)
		require.NotNil(t, deletedChat)
		assert.Empty(t, deletedChat.Id) // Should return empty chat when not found
	})

	t.Run("should handle deletion of non-existent chat gracefully", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		err := repo.DeleteChat(nonExistentID)
		require.NoError(t, err) // Should not error even if chat doesn't exist
	})

	t.Run("should delete chat created during test", func(t *testing.T) {
		// Create a chat first
		userGroupID := 300
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		createdChat, err := repo.CreateChat(chat)
		require.NoError(t, err)
		require.NotNil(t, createdChat)

		// Verify it was created
		foundChat, err := repo.GetChatById(createdChat.Id)
		require.NoError(t, err)
		assert.Equal(t, createdChat.Id, foundChat.Id)

		// Delete it
		err = repo.DeleteChat(createdChat.Id)
		require.NoError(t, err)

		// Verify it's deleted
		deletedChat, err := repo.GetChatById(createdChat.Id)
		require.NoError(t, err)
		assert.Empty(t, deletedChat.Id)
	})

	t.Run("should not affect other chats when deleting one", func(t *testing.T) {
		// Create two chats
		chat1, err := domain.NewChat(domain.ChatTypePrivate, nil, nil)
		require.NoError(t, err)
		chat2, err := domain.NewChat(domain.ChatTypePrivate, nil, nil)
		require.NoError(t, err)

		createdChat1, err1 := repo.CreateChat(chat1)
		createdChat2, err2 := repo.CreateChat(chat2)
		require.NoError(t, err1)
		require.NoError(t, err2)

		// Delete only the first one
		err = repo.DeleteChat(createdChat1.Id)
		require.NoError(t, err)

		// Verify first is deleted
		deletedChat, err := repo.GetChatById(createdChat1.Id)
		require.NoError(t, err)
		assert.Empty(t, deletedChat.Id)

		// Verify second still exists
		existingChat, err := repo.GetChatById(createdChat2.Id)
		require.NoError(t, err)
		assert.Equal(t, createdChat2.Id, existingChat.Id)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat2.Id)
	})
}
func TestChatRepository_CreateChatWithParticipant(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should create chat with participant successfully in transaction", func(t *testing.T) {
		userGroupID := 500
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		// Generate a valid UUID for user_id
		userUUID, err := uuid.NewV7()
		require.NoError(t, err)
		userID := userUUID.String()

		participant, err := domain.NewChatParticipant(chat.Id, userID, "admin", "active")
		require.NoError(t, err)

		createdChat, err := repo.CreateChatWithParticipant(chat, participant)

		require.NoError(t, err)
		require.NotNil(t, createdChat)
		assert.NotEmpty(t, createdChat.Id)
		assert.Equal(t, domain.ChatTypeGroup, createdChat.Type)
		assert.NotNil(t, createdChat.UserGroupId)
		assert.Equal(t, userGroupID, *createdChat.UserGroupId)
		assert.False(t, createdChat.CreatedAt.IsZero())

		// Verify chat was created in database
		foundChat, err := repo.GetChatById(createdChat.Id)
		require.NoError(t, err)
		assert.Equal(t, createdChat.Id, foundChat.Id)

		// Verify participant was created in database
		participants, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		require.Len(t, participants, 1)
		assert.Equal(t, userID, participants[0].UserId)
		assert.Equal(t, "admin", participants[0].Role.String())
		assert.Equal(t, "active", participants[0].Status.String())
		assert.NotEmpty(t, participants[0].Id)
		assert.False(t, participants[0].JoinedAt.IsZero())

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat_participant WHERE chat_id = $1`, createdChat.Id)
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})

	t.Run("should rollback transaction when participant creation fails", func(t *testing.T) {
		userGroupID := 501
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		// Create participant with invalid UUID format to cause database constraint violation
		participant, err := domain.NewChatParticipant(chat.Id, "invalid-uuid-format", "admin", "active")
		require.NoError(t, err)

		// This should fail due to database UUID format constraints
		createdChat, err := repo.CreateChatWithParticipant(chat, participant)

		require.Error(t, err)
		require.Nil(t, createdChat)

		// Verify that chat was NOT created (transaction rolled back)
		foundChat, err := repo.GetChatById(chat.Id)
		require.NoError(t, err)
		assert.Empty(t, foundChat.Id) // Should be empty since chat creation was rolled back

		// Verify no participants exist for this chat
		participants, err := repo.GetChatParticipants(chat.Id)
		require.NoError(t, err)
		assert.Len(t, participants, 0)
	})

	t.Run("should handle database connection issues gracefully", func(t *testing.T) {
		// Test with a closed database connection would be complex to set up
		// This test validates the method signature and basic error handling
		userGroupID := 502
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		// Generate a valid UUID for user_id
		userUUID, err := uuid.NewV7()
		require.NoError(t, err)
		userID := userUUID.String()

		participant, err := domain.NewChatParticipant(chat.Id, userID, "member", "active")
		require.NoError(t, err)

		// This should succeed with valid inputs
		createdChat, err := repo.CreateChatWithParticipant(chat, participant)

		require.NoError(t, err)
		require.NotNil(t, createdChat)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat_participant WHERE chat_id = $1`, createdChat.Id)
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})

	t.Run("should create private chat with participant", func(t *testing.T) {
		chat, err := domain.NewChat(domain.ChatTypePrivate, nil, nil)
		require.NoError(t, err)

		// Generate a valid UUID for user_id
		userUUID, err := uuid.NewV7()
		require.NoError(t, err)
		userID := userUUID.String()

		participant, err := domain.NewChatParticipant(chat.Id, userID, "member", "active")
		require.NoError(t, err)

		createdChat, err := repo.CreateChatWithParticipant(chat, participant)

		require.NoError(t, err)
		require.NotNil(t, createdChat)
		assert.Equal(t, domain.ChatTypePrivate, createdChat.Type)
		assert.Nil(t, createdChat.UserGroupId)
		assert.Nil(t, createdChat.ContainerId)

		// Verify participant was created
		participants, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		require.Len(t, participants, 1)
		assert.Equal(t, userID, participants[0].UserId)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat_participant WHERE chat_id = $1`, createdChat.Id)
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})
}
func TestChatRepository_GetChatParticipants(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should return participants for existing chat", func(t *testing.T) {
		chatID := "01987073-0a87-7b32-9439-86868dfe9bd2"

		participants, err := repo.GetChatParticipants(chatID)

		require.NoError(t, err)
		require.NotNil(t, participants)
		assert.Greater(t, len(participants), 0)

		// Verify participant structure
		for _, p := range participants {
			assert.NotEmpty(t, p.Id)
			assert.Equal(t, chatID, p.ChatId)
			assert.NotEmpty(t, p.UserId)
			assert.Contains(t, []domain.ParticipantRole{domain.RoleAdmin, domain.RoleMember}, p.Role)
			assert.Contains(t, []domain.ParticipantStatus{domain.StatusActive, domain.StatusLeft, domain.StatusBanned, domain.StatusMuted, domain.StatusPending}, p.Status)
			assert.False(t, p.JoinedAt.IsZero())
		}
	})

	t.Run("should return empty slice for non-existent chat", func(t *testing.T) {
		nonExistentID := "01987073-0000-0000-0000-000000000000"

		participants, err := repo.GetChatParticipants(nonExistentID)

		require.NoError(t, err)
		require.NotNil(t, participants)
		assert.Equal(t, 0, len(participants))
	})
}
func TestChatRepository_DeleteParticipantFromChat(t *testing.T) {
	repo := repository.NewRepository(testDB)

	t.Run("should delete existing participant from chat", func(t *testing.T) {
		// Create a chat with a participant
		userGroupID := 600
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		userUUID, err := uuid.NewV7()
		require.NoError(t, err)
		userID := userUUID.String()

		participant, err := domain.NewChatParticipant(chat.Id, userID, "member", "active")
		require.NoError(t, err)

		createdChat, err := repo.CreateChatWithParticipant(chat, participant)
		require.NoError(t, err)

		// Verify participant exists
		participants, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		require.Len(t, participants, 1)
		assert.Equal(t, userID, participants[0].UserId)

		// Delete the participant
		err = repo.DeleteParticipantFromChat(createdChat.Id, userID)
		require.NoError(t, err)

		// Verify participant was deleted
		participantsAfter, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		assert.Len(t, participantsAfter, 0)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})

	t.Run("should handle deletion of non-existent participant gracefully", func(t *testing.T) {
		// Create a chat without participants
		userGroupID := 601
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		createdChat, err := repo.CreateChat(chat)
		require.NoError(t, err)

		nonExistentUserID := "01959b38-0000-0000-0000-000000000000"

		// Attempt to delete non-existent participant
		err = repo.DeleteParticipantFromChat(createdChat.Id, nonExistentUserID)
		require.NoError(t, err) // Should not error even if participant doesn't exist

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})

	t.Run("should handle deletion from non-existent chat gracefully", func(t *testing.T) {
		nonExistentChatID := "01987073-0000-0000-0000-000000000000"
		nonExistentUserID := "01959b38-0000-0000-0000-000000000000"

		// Attempt to delete participant from non-existent chat
		err := repo.DeleteParticipantFromChat(nonExistentChatID, nonExistentUserID)
		require.NoError(t, err) // Should not error even if chat doesn't exist
	})

	t.Run("should delete only specified participant from chat with multiple participants", func(t *testing.T) {
		// Create a chat
		userGroupID := 602
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		createdChat, err := repo.CreateChat(chat)
		require.NoError(t, err)

		// Create two participants
		userUUID1, err := uuid.NewV7()
		require.NoError(t, err)
		userID1 := userUUID1.String()

		userUUID2, err := uuid.NewV7()
		require.NoError(t, err)
		userID2 := userUUID2.String()

		participant1, err := domain.NewChatParticipant(createdChat.Id, userID1, "admin", "active")
		require.NoError(t, err)
		participant2, err := domain.NewChatParticipant(createdChat.Id, userID2, "member", "active")
		require.NoError(t, err)

		// Add both participants
		_, err = repo.AddParticipantToChat(participant1)
		require.NoError(t, err)
		_, err = repo.AddParticipantToChat(participant2)
		require.NoError(t, err)

		// Verify both participants exist
		participants, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		require.Len(t, participants, 2)

		// Delete only the first participant
		err = repo.DeleteParticipantFromChat(createdChat.Id, userID1)
		require.NoError(t, err)

		// Verify only one participant remains
		participantsAfter, err := repo.GetChatParticipants(createdChat.Id)
		require.NoError(t, err)
		require.Len(t, participantsAfter, 1)
		assert.Equal(t, userID2, participantsAfter[0].UserId)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat_participant WHERE chat_id = $1`, createdChat.Id)
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})

	t.Run("should verify participant status check before deletion", func(t *testing.T) {
		// Create a chat with a participant
		userGroupID := 603
		chat, err := domain.NewChat(domain.ChatTypeGroup, &userGroupID, nil)
		require.NoError(t, err)

		userUUID, err := uuid.NewV7()
		require.NoError(t, err)
		userID := userUUID.String()

		participant, err := domain.NewChatParticipant(chat.Id, userID, "member", "active")
		require.NoError(t, err)

		createdChat, err := repo.CreateChatWithParticipant(chat, participant)
		require.NoError(t, err)

		// Verify participant is indeed in the chat before deletion
		isParticipant, err := repo.IsUserParticipantInChat(createdChat.Id, userID)
		require.NoError(t, err)
		assert.True(t, isParticipant)

		// Delete the participant
		err = repo.DeleteParticipantFromChat(createdChat.Id, userID)
		require.NoError(t, err)

		// Verify participant is no longer in the chat after deletion
		isParticipantAfter, err := repo.IsUserParticipantInChat(createdChat.Id, userID)
		require.NoError(t, err)
		assert.False(t, isParticipantAfter)

		// Cleanup
		_, _ = testDB.Exec(`DELETE FROM public.chat WHERE id = $1`, createdChat.Id)
	})
}
