package match

import (
	"church-match-api/internal/profile"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	GetCandidates(ctx context.Context, userID string) ([]profile.Profile, error)
}

type matchRepository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &matchRepository{
		db: db,
	}
}

func (r *matchRepository) GetCandidates(ctx context.Context, userID string) ([]profile.Profile, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 1. Get IDs of users to exclude (already sent or received requests)
	requestCursor, err := r.db.Collection("requests").Find(ctx, bson.M{
		"$or": []bson.M{
			{"sender_id": uID},
			{"receiver_id": uID},
		},
	})
	if err != nil {
		return nil, err
	}
	defer requestCursor.Close(ctx)

	excludeIDs := []primitive.ObjectID{uID}
	for requestCursor.Next(ctx) {
		var req struct {
			SenderID   primitive.ObjectID `bson:"sender_id"`
			ReceiverID primitive.ObjectID `bson:"receiver_id"`
		}
		if err := requestCursor.Decode(&req); err == nil {
			if req.SenderID == uID {
				excludeIDs = append(excludeIDs, req.ReceiverID)
			} else {
				excludeIDs = append(excludeIDs, req.SenderID)
			}
		}
	}

	// 2. Fetch active users who have profiles and are NOT in excludeIDs
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": bson.M{"$nin": excludeIDs},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user_info",
		}}},
		{{Key: "$unwind", Value: "$user_info"}},
		{{Key: "$match", Value: bson.M{
			"user_info.status": "active",
		}}},
	}

	cursor, err := r.db.Collection("profiles").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var candidates []profile.Profile
	if err := cursor.All(ctx, &candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}
