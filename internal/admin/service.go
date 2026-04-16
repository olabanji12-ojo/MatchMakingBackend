package admin

import (
	"context"
)

type Service interface {
	GetUsers(ctx context.Context, status string, page, limit int) ([]UserWithProfile, error)
	ApproveUser(ctx context.Context, id string) error
	RejectUser(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error
	GetStats(ctx context.Context) (Stats, error)
}

type adminService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &adminService{
		repo: repo,
	}
}

func (s *adminService) GetUsers(ctx context.Context, status string, page, limit int) ([]UserWithProfile, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.repo.GetUsers(ctx, status, page, limit)
}

func (s *adminService) ApproveUser(ctx context.Context, id string) error {
	return s.repo.UpdateUserStatus(ctx, id, "active")
}

func (s *adminService) RejectUser(ctx context.Context, id string) error {
	return s.repo.UpdateUserStatus(ctx, id, "rejected")
}

func (s *adminService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *adminService) GetStats(ctx context.Context) (Stats, error) {
	return s.repo.GetStats(ctx)
}
