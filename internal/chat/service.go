package chat

import (
	"church-match-api/internal/profile"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service interface {
	GetChatList(ctx context.Context, userID string) ([]ChatSummary, error)
	GetMessages(ctx context.Context, userID, chatID string, limit int, before string) ([]Message, error)
	SendMessage(ctx context.Context, userID, chatID, content string) (*Message, error)
	GetChatWithUser(ctx context.Context, userID, otherUserID string) (*Chat, error)
	VerifyParticipation(ctx context.Context, userID, chatID string) (string, error) // Returns other user's ID
}

type chatService struct {
	repo        Repository
	profileRepo profile.Repository
}

func NewService(repo Repository, profileRepo profile.Repository) Service {
	return &chatService{
		repo:        repo,
		profileRepo: profileRepo,
	}
}

func (s *chatService) GetChatList(ctx context.Context, userID string) ([]ChatSummary, error) {
	chats, err := s.repo.GetChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	summaries := []ChatSummary{}
	for _, c := range chats {
		otherUserID := c.User1ID.Hex()
		if otherUserID == userID {
			otherUserID = c.User2ID.Hex()
		}

		publicProf, _ := s.profileRepo.FindByUserID(ctx, otherUserID)
		lastMsg, _ := s.repo.GetLastMessage(ctx, c.ID)

		summary := ChatSummary{
			Chat: c,
		}
		if publicProf != nil {
			summary.OtherUser = profile.PublicProfile{
				Name:     publicProf.Name,
				Age:      publicProf.Age,
				Gender:   publicProf.Gender,
				Church:   publicProf.Church,
				Location: publicProf.Location,
				Values:   publicProf.Values,
			}
		}
		summary.LastMessage = lastMsg
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (s *chatService) GetMessages(ctx context.Context, userID, chatID string, limit int, before string) ([]Message, error) {
	_, err := s.VerifyParticipation(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	if limit < 1 {
		limit = 50
	}
	return s.repo.GetMessages(ctx, chatID, limit, before)
}

func (s *chatService) SendMessage(ctx context.Context, userID, chatID, content string) (*Message, error) {
	_, err := s.VerifyParticipation(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	cID, _ := primitive.ObjectIDFromHex(chatID)
	sID, _ := primitive.ObjectIDFromHex(userID)

	msg := &Message{
		ChatID:   cID,
		SenderID: sID,
		Content:  content,
	}

	err = s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *chatService) GetChatWithUser(ctx context.Context, userID, otherUserID string) (*Chat, error) {
	return s.repo.GetChatWithUser(ctx, userID, otherUserID)
}

func (s *chatService) VerifyParticipation(ctx context.Context, userID, chatID string) (string, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return "", errors.New("chat not found")
	}

	if chat.User1ID.Hex() == userID {
		return chat.User2ID.Hex(), nil
	}
	if chat.User2ID.Hex() == userID {
		return chat.User1ID.Hex(), nil
	}

	return "", errors.New("not a participant in this chat")
}
