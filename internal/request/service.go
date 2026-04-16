package request

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service interface {
	SendRequest(ctx context.Context, senderID, receiverID string) error
	GetReceived(ctx context.Context, userID string) ([]Request, error)
	GetSent(ctx context.Context, userID string) ([]Request, error)
	AcceptRequest(ctx context.Context, userID, requestID string) (string, error)
	RejectRequest(ctx context.Context, userID, requestID string) error
	CancelRequest(ctx context.Context, userID, requestID string) error
}

type requestService struct {
	repo        Repository
	redisClient *redis.Client
}

func NewService(repo Repository, redisClient *redis.Client) Service {
	return &requestService{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *requestService) SendRequest(ctx context.Context, senderID, receiverID string) error {
	if senderID == receiverID {
		return errors.New("cannot match with yourself")
	}

	existing, err := s.repo.FindExisting(ctx, senderID, receiverID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("request already exists")
	}

	sID, _ := primitive.ObjectIDFromHex(senderID)
	rID, _ := primitive.ObjectIDFromHex(receiverID)

	req := &Request{
		ID:         primitive.NewObjectID(),
		SenderID:   sID,
		ReceiverID: rID,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.repo.CreateRequest(ctx, req)
	if err != nil {
		return err
	}

	// Invalidate cache for both users
	s.redisClient.Del(ctx, "matches:"+senderID)
	s.redisClient.Del(ctx, "matches:"+receiverID)

	return nil
}

func (s *requestService) GetReceived(ctx context.Context, userID string) ([]Request, error) {
	return s.repo.GetReceived(ctx, userID)
}

func (s *requestService) GetSent(ctx context.Context, userID string) ([]Request, error) {
	return s.repo.GetSent(ctx, userID)
}

func (s *requestService) AcceptRequest(ctx context.Context, userID, requestID string) (string, error) {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil || req == nil {
		return "", errors.New("request not found")
	}

	if req.ReceiverID.Hex() != userID {
		return "", errors.New("not authorized to accept this request")
	}

	if req.Status != "pending" {
		return "", errors.New("request already processed")
	}

	err = s.repo.UpdateStatus(ctx, requestID, "accepted")
	if err != nil {
		return "", err
	}

	// Create Chat
	chatID, err := s.repo.CreateChat(ctx, req.SenderID, req.ReceiverID)
	return chatID, err
}

func (s *requestService) RejectRequest(ctx context.Context, userID, requestID string) error {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil || req == nil {
		return errors.New("request not found")
	}

	if req.ReceiverID.Hex() != userID {
		return errors.New("not authorized to reject this request")
	}

	if req.Status != "pending" {
		return errors.New("request already processed")
	}

	return s.repo.UpdateStatus(ctx, requestID, "rejected")
}

func (s *requestService) CancelRequest(ctx context.Context, userID, requestID string) error {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil || req == nil {
		return errors.New("request not found")
	}

	if req.SenderID.Hex() != userID {
		return errors.New("not authorized to cancel this request")
	}

	if req.Status != "pending" {
		return errors.New("cannot cancel a processed request")
	}

	return s.repo.DeleteRequest(ctx, requestID)
}
