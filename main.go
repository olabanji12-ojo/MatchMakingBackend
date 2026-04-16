package main

import (
	"church-match-api/config"
	"church-match-api/internal/admin"
	"church-match-api/internal/auth"
	"church-match-api/internal/chat"
	"church-match-api/internal/match"
	"church-match-api/internal/profile"
	"church-match-api/internal/request"
	"church-match-api/pkg/database"
	"church-match-api/pkg/token"
	"church-match-api/router"
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Connect to MongoDB
	mongoClient := database.ConnectMongo(cfg.MongoURI)
	db := mongoClient.Database(cfg.MongoDB)

	// 3. Connect to Redis
	redisClient := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)

	// 4. Create Indexes
	createIndexes(db)

	// 5. Initialize Services/Handlers
	jwtService := token.NewJWTService(cfg.JWTSecret)
	hub := chat.NewHub()
	go hub.Run()

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtService, redisClient, cfg.BcryptCost, cfg.JWTExpiryHours, cfg.AdminRegistrationSecret)
	authHandler := auth.NewHandler(authService)

	profileRepo := profile.NewRepository(db)
	profileService := profile.NewService(profileRepo)
	profileHandler := profile.NewHandler(profileService)

	adminRepo := admin.NewRepository(db)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	matchRepo := match.NewRepository(db)
	matchService := match.NewService(matchRepo, profileRepo, redisClient)
	matchHandler := match.NewHandler(matchService)

	requestRepo := request.NewRepository(db)
	requestService := request.NewService(requestRepo, redisClient)
	requestHandler := request.NewHandler(requestService)

	chatRepo := chat.NewRepository(db)
	chatService := chat.NewService(chatRepo, profileRepo)
	chatHandler := chat.NewHandler(chatService, hub)

	// 6. Setup Router
	r := router.SetupRouter(
		authHandler,
		profileHandler,
		matchHandler,
		requestHandler,
		chatHandler,
		adminHandler,
		jwtService,
		redisClient,
		cfg.AllowedOrigins,
	)

	// 7. Start Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func createIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Users: Email unique
	_, _ = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Profiles: UserID unique
	_, _ = db.Collection("profiles").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Requests: Compound unique
	_, _ = db.Collection("requests").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sender_id", Value: 1}, {Key: "receiver_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Messages: ChatID index
	_, _ = db.Collection("messages").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "chat_id", Value: 1}},
	})
}
