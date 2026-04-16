package admin

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	GetUsers(ctx context.Context, status string, page, limit int) ([]UserWithProfile, error)
	UpdateUserStatus(ctx context.Context, id string, status string) error
	DeleteUser(ctx context.Context, id string) error
	GetStats(ctx context.Context) (Stats, error)
}

type adminRepository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &adminRepository{
		db: db,
	}
}

func (r *adminRepository) GetUsers(ctx context.Context, status string, page, limit int) ([]UserWithProfile, error) {
	match := bson.M{}
	if status != "" {
		match["status"] = status
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "profiles",
			"localField":   "_id",
			"foreignField": "user_id",
			"as":           "profile",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$profile",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$skip", Value: int64((page - 1) * limit)}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []UserWithProfile
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *adminRepository) UpdateUserStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	res, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *adminRepository) DeleteUser(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Transactions would be ideal but for simplicity we do sequential deletes
	_, err = r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	_, _ = r.db.Collection("profiles").DeleteOne(ctx, bson.M{"user_id": objID})
	_, _ = r.db.Collection("requests").DeleteMany(ctx, bson.M{
		"$or": []bson.M{
			{"sender_id": objID},
			{"receiver_id": objID},
		},
	})

	return nil
}

func (r *adminRepository) GetStats(ctx context.Context) (Stats, error) {
	var stats Stats

	totalUsers, _ := r.db.Collection("users").CountDocuments(ctx, bson.M{})
	pendingUsers, _ := r.db.Collection("users").CountDocuments(ctx, bson.M{"status": "pending"})
	activeUsers, _ := r.db.Collection("users").CountDocuments(ctx, bson.M{"status": "active"})
	rejectedUsers, _ := r.db.Collection("users").CountDocuments(ctx, bson.M{"status": "rejected"})

	totalRequests, _ := r.db.Collection("requests").CountDocuments(ctx, bson.M{})
	acceptedRequests, _ := r.db.Collection("requests").CountDocuments(ctx, bson.M{"status": "accepted"})

	totalChats, _ := r.db.Collection("chats").CountDocuments(ctx, bson.M{})
	totalMessages, _ := r.db.Collection("messages").CountDocuments(ctx, bson.M{})

	stats.TotalUsers = int(totalUsers)
	stats.PendingUsers = int(pendingUsers)
	stats.ActiveUsers = int(activeUsers)
	stats.RejectedUsers = int(rejectedUsers)
	stats.TotalRequests = int(totalRequests)
	stats.AcceptedRequests = int(acceptedRequests)
	stats.TotalChats = int(totalChats)
	stats.TotalMessages = int(totalMessages)

	return stats, nil
}
