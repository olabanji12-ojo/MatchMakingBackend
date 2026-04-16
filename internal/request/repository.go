package request

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	CreateRequest(ctx context.Context, req *Request) error
	FindByID(ctx context.Context, id string) (*Request, error)
	FindExisting(ctx context.Context, senderID, receiverID string) (*Request, error)
	GetReceived(ctx context.Context, userID string) ([]Request, error)
	GetSent(ctx context.Context, userID string) ([]Request, error)
	UpdateStatus(ctx context.Context, id, status string) error
	DeleteRequest(ctx context.Context, id string) error
	CreateChat(ctx context.Context, user1ID, user2ID primitive.ObjectID) (string, error)
}

type requestRepository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &requestRepository{
		db: db,
	}
}

func (r *requestRepository) CreateRequest(ctx context.Context, req *Request) error {
	_, err := r.db.Collection("requests").InsertOne(ctx, req)
	return err
}

func (r *requestRepository) FindByID(ctx context.Context, id string) (*Request, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var req Request
	err = r.db.Collection("requests").FindOne(ctx, bson.M{"_id": objID}).Decode(&req)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *requestRepository) FindExisting(ctx context.Context, senderID, receiverID string) (*Request, error) {
	sID, _ := primitive.ObjectIDFromHex(senderID)
	rID, _ := primitive.ObjectIDFromHex(receiverID)

	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": sID, "receiver_id": rID},
			{"sender_id": rID, "receiver_id": sID},
		},
	}

	var req Request
	err := r.db.Collection("requests").FindOne(ctx, filter).Decode(&req)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *requestRepository) GetReceived(ctx context.Context, userID string) ([]Request, error) {
	uID, _ := primitive.ObjectIDFromHex(userID)
	cursor, err := r.db.Collection("requests").Find(ctx, bson.M{"receiver_id": uID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *requestRepository) GetSent(ctx context.Context, userID string) ([]Request, error) {
	uID, _ := primitive.ObjectIDFromHex(userID)
	cursor, err := r.db.Collection("requests").Find(ctx, bson.M{"sender_id": uID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []Request
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *requestRepository) UpdateStatus(ctx context.Context, id, status string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.db.Collection("requests").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
	return err
}

func (r *requestRepository) DeleteRequest(ctx context.Context, id string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.db.Collection("requests").DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *requestRepository) CreateChat(ctx context.Context, user1ID, user2ID primitive.ObjectID) (string, error) {
	chat := bson.M{
		"user1_id":   user1ID,
		"user2_id":   user2ID,
		"created_at": time.Now(),
	}

	res, err := r.db.Collection("chats").InsertOne(ctx, chat)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}
