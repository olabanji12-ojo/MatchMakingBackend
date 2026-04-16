package match

import (
	"church-match-api/internal/profile"
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	GetMatches(ctx context.Context, userID string, page, limit int) ([]MatchResult, error)
}

type matchService struct {
	repo        Repository
	profileRepo profile.Repository
	redisClient *redis.Client
}

func NewService(repo Repository, profileRepo profile.Repository, redisClient *redis.Client) Service {
	return &matchService{
		repo:        repo,
		profileRepo: profileRepo,
		redisClient: redisClient,
	}
}

func (s *matchService) GetMatches(ctx context.Context, userID string, page, limit int) ([]MatchResult, error) {
	// 1. Try Cache
	cacheKey := "matches:" + userID
	if data, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResults []MatchResult
		if err := json.Unmarshal([]byte(data), &cachedResults); err == nil {
			return s.paginate(cachedResults, page, limit), nil
		}
	}

	// 2. Get current user profile
	currentUserProfile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil || currentUserProfile == nil {
		return nil, err
	}

	// 3. Get candidates
	candidates, err := s.repo.GetCandidates(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 4. Calculate scores
	var results []MatchResult
	for _, cand := range candidates {
		score := s.calculateScore(currentUserProfile, &cand)
		if score < 2 {
			continue
		}
		results = append(results, MatchResult{
			UserID: cand.UserID.Hex(),
			Score:  score,
			Profile: profile.PublicProfile{
				Name:     cand.Name,
				Age:      cand.Age,
				Gender:   cand.Gender,
				Church:   cand.Church,
				Location: cand.Location,
				Values:   cand.Values,
			},
		})
	}

	// 5. Sort Score Descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 6. Cache full results (10m)
	cachedData, _ := json.Marshal(results)
	_ = s.redisClient.Set(ctx, cacheKey, cachedData, 10*time.Minute).Err()

	return s.paginate(results, page, limit), nil
}

func (s *matchService) calculateScore(current, cand *profile.Profile) int {
	score := 0

	// Same Church (+3)
	if current.Church == cand.Church {
		score += 3
	}

	// Shared Values (+1 each)
	valuesMap := make(map[string]bool)
	for _, v := range current.Values {
		valuesMap[v] = true
	}
	for _, v := range cand.Values {
		if valuesMap[v] {
			score += 1
		}
	}

	// Gender Preference Match (+2): candidate's gender matches what current user prefers
	if current.PreferredGender == "any" || current.PreferredGender == cand.Gender {
		score += 2
	}

	// Mutual Interest (+1): candidate also prefers current user's gender
	if cand.PreferredGender == "any" || cand.PreferredGender == current.Gender {
		score += 1
	}

	// Age Range Compatibility (+1 each direction)
	if current.MinAge > 0 && current.MaxAge > 0 {
		if cand.Age >= current.MinAge && cand.Age <= current.MaxAge {
			score += 1
		}
	}
	if cand.MinAge > 0 && cand.MaxAge > 0 {
		if current.Age >= cand.MinAge && current.Age <= cand.MaxAge {
			score += 1
		}
	}

	return score
}

func (s *matchService) paginate(results []MatchResult, page, limit int) []MatchResult {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	start := (page - 1) * limit
	if start >= len(results) {
		return []MatchResult{}
	}

	end := start + limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}
