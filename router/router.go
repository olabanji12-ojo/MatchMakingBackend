package router

import (
	"church-match-api/internal/admin"
	"church-match-api/internal/auth"
	"church-match-api/internal/chat"
	"church-match-api/internal/match"
	"church-match-api/internal/profile"
	"church-match-api/internal/request"
	"church-match-api/pkg/middleware"
	"church-match-api/pkg/token"
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(
	authHandler *auth.Handler,
	profileHandler *profile.Handler,
	matchHandler *match.Handler,
	requestHandler *request.Handler,
	chatHandler *chat.Handler,
	adminHandler *admin.Handler,
	jwtService *token.JWTService,
	redisClient *redis.Client,
) http.Handler {
	r := mux.NewRouter()

	v1 := r.PathPrefix("/api/v1").Subrouter()

	// AUTH (public)
	v1.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	v1.HandleFunc("/auth/admin/register", authHandler.RegisterAdmin).Methods("POST")
	v1.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	v1.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	// Protected Routes
	protected := v1.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtService, redisClient))

	// AUTH (protected)
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET")

	// PROFILE
	protected.HandleFunc("/users/me/profile", profileHandler.UpdateMyProfile).Methods("PUT")
	protected.HandleFunc("/users/me", profileHandler.GetMyProfile).Methods("GET")
	protected.HandleFunc("/users/{id}/public", profileHandler.GetPublicProfile).Methods("GET")

	// MATCHING
	protected.HandleFunc("/matches", matchHandler.GetMatches).Methods("GET")

	// REQUESTS
	protected.HandleFunc("/requests", requestHandler.SendRequest).Methods("POST")
	protected.HandleFunc("/requests/received", requestHandler.GetReceived).Methods("GET")
	protected.HandleFunc("/requests/sent", requestHandler.GetSent).Methods("GET")
	protected.HandleFunc("/requests/{id}/accept", requestHandler.AcceptRequest).Methods("PUT")
	protected.HandleFunc("/requests/{id}/reject", requestHandler.RejectRequest).Methods("PUT")
	protected.HandleFunc("/requests/{id}", requestHandler.CancelRequest).Methods("DELETE")

	// CHAT (REST)
	protected.HandleFunc("/chats", chatHandler.GetChats).Methods("GET")
	protected.HandleFunc("/chats/{id}/messages", chatHandler.GetMessages).Methods("GET")
	protected.HandleFunc("/chats/{id}/messages", chatHandler.SendMessage).Methods("POST")

	// WebSocket: Uses query param ?token= because browsers cannot send custom headers during WS upgrade
	v1.HandleFunc("/ws/chat", func(w http.ResponseWriter, req *http.Request) {
		tokenStr := req.URL.Query().Get("token")
		if tokenStr == "" {
			// Fallback: check Authorization header
			authHeader := req.Header.Get("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				tokenStr = parts[1]
			}
		}
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := jwtService.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, middleware.RoleKey, claims.Role)
		chatHandler.ServeWS(w, req.WithContext(ctx))
	})

	// ADMIN
	adminOnly := protected.PathPrefix("/admin").Subrouter()
	adminOnly.Use(middleware.AdminMiddleware)
	adminOnly.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	adminOnly.HandleFunc("/users/{id}/approve", adminHandler.ApproveUser).Methods("PUT")
	adminOnly.HandleFunc("/users/{id}/reject", adminHandler.RejectUser).Methods("PUT")
	adminOnly.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	adminOnly.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	// Global CORS Wrapper
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		r.ServeHTTP(w, req)
	})
}
