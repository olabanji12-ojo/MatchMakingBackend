package admin

import (
	"church-match-api/internal/auth"
	"church-match-api/internal/profile"
)

type UserWithProfile struct {
	auth.User       `bson:",inline"`
	Profile *profile.Profile `json:"profile,omitempty"`
}

type Stats struct {
	TotalUsers       int `json:"total_users"`
	PendingUsers     int `json:"pending_users"`
	ActiveUsers      int `json:"active_users"`
	RejectedUsers    int `json:"rejected_users"`
	TotalRequests    int `json:"total_requests"`
	AcceptedRequests int `json:"accepted_requests"`
	TotalChats       int `json:"total_chats"`
	TotalMessages    int `json:"total_messages"`
}
