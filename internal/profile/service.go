package profile

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service interface {
	UpdateProfile(ctx context.Context, userID string, profile Profile) (*Profile, error)
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	GetPublicProfile(ctx context.Context, userID string) (*PublicProfile, error)
}

type profileService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &profileService{
		repo: repo,
	}
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, profile Profile) (*Profile, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	profile.UserID = uID
	profile.UpdatedAt = time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	err = s.repo.UpsertProfile(ctx, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *profileService) GetPublicProfile(ctx context.Context, userID string) (*PublicProfile, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}

	return &PublicProfile{
		Name:     profile.Name,
		Age:      profile.Age,
		Gender:   profile.Gender,
		Church:   profile.Church,
		Location: profile.Location,
		Values:   profile.Values,
	}, nil
}
