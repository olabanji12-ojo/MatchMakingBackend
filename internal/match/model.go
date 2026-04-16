package match

import "church-match-api/internal/profile"

type MatchResult struct {
	Profile profile.PublicProfile `json:"profile"`
	UserID  string                `json:"user_id"`
	Score   int                   `json:"score"`
}
