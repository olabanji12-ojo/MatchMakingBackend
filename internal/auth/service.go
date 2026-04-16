package auth

import (
	"church-match-api/pkg/token"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) error
	RegisterAdmin(ctx context.Context, req AdminRegisterRequest, secret string) error
	Login(ctx context.Context, req LoginRequest) (AuthResponse, error)
	Logout(ctx context.Context, userID string) error
	GetMe(ctx context.Context, userID string) (*User, error)
}

type authService struct {
	repo        Repository
	jwtService  *token.JWTService
	redisClient *redis.Client
	bcryptCost  int
	jwtExpiry   int
	adminSecret string
}

func NewService(repo Repository, jwtService *token.JWTService, redisClient *redis.Client, bcryptCost, jwtExpiry int, adminSecret string) Service {
	return &authService{
		repo:        repo,
		jwtService:  jwtService,
		redisClient: redisClient,
		bcryptCost:  bcryptCost,
		jwtExpiry:   jwtExpiry,
		adminSecret: adminSecret,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) error {
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return err
	}

	user := &User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *authService) RegisterAdmin(ctx context.Context, req AdminRegisterRequest, secret string) error {
	if secret != s.adminSecret {
		return errors.New("invalid admin secret")
	}

	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return err
	}

	user := &User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "admin",
		Status:    "approved",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return AuthResponse{}, err
	}
	if user == nil {
		return AuthResponse{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return AuthResponse{}, errors.New("invalid credentials")
	}

	if user.Status == "rejected" {
		return AuthResponse{}, errors.New("account rejected. Please contact support")
	}

	// Pending users ARE allowed to log in. The frontend redirects them to the Pending Dashboard.
	tokenStr, err := s.jwtService.GenerateToken(user.ID.Hex(), user.Role, s.jwtExpiry)
	if err != nil {
		return AuthResponse{}, err
	}

	// Store session in Redis
	err = s.redisClient.Set(ctx, "session:"+user.ID.Hex(), tokenStr, time.Duration(s.jwtExpiry)*time.Hour).Err()
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: tokenStr,
		User:  *user,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	return s.redisClient.Del(ctx, "session:"+userID).Err()
}

func (s *authService) GetMe(ctx context.Context, userID string) (*User, error) {
	return s.repo.FindByID(ctx, userID)
}
