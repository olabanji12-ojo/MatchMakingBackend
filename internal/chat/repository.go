package chat

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	GetChats(ctx context.Context, userID string) ([]Chat, error)
	GetChatByID(ctx context.Context, id string) (*Chat, error)
	GetMessages(ctx context.Context, chatID string, limit int, before string) ([]Message, error)
	GetLastMessage(ctx context.Context, chatID primitive.ObjectID) (*Message, error)
	SaveMessage(ctx context.Context, msg *Message) error
	GetChatWithUser(ctx context.Context, user1ID, user2ID string) (*Chat, error)
}

type chatRepository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &chatRepository{
		db: db,
	}
}

func (r *chatRepository) GetChats(ctx context.Context, userID string) ([]Chat, error) {
	uID, _ := primitive.ObjectIDFromHex(userID)
	cursor, err := r.db.Collection("chats").Find(ctx, bson.M{
		"$or": []bson.M{
			{"user1_id": uID},
			{"user2_id": uID},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []Chat
	if err := cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepository) GetChatByID(ctx context.Context, id string) (*Chat, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var chat Chat
	err = r.db.Collection("chats").FindOne(ctx, bson.M{"_id": objID}).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) GetMessages(ctx context.Context, chatID string, limit int, before string) ([]Message, error) {
	cID, _ := primitive.ObjectIDFromHex(chatID)
	filter := bson.M{"chat_id": cID}
	if before != "" {
		bID, _ := primitive.ObjectIDFromHex(before)
		filter["_id"] = bson.M{"$lt": bID}
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.M{"_id": -1})
	cursor, err := r.db.Collection("messages").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *chatRepository) GetLastMessage(ctx context.Context, chatID primitive.ObjectID) (*Message, error) {
	opts := options.FindOne().SetSort(bson.M{"_id": -1})
	var msg Message
	err := r.db.Collection("messages").FindOne(ctx, bson.M{"chat_id": chatID}, opts).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *Message) error {
	msg.CreatedAt = time.Now()
	res, err := r.db.Collection("messages").InsertOne(ctx, msg)
	if err != nil {
		return err
	}
	msg.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *chatRepository) GetChatWithUser(ctx context.Context, user1ID, user2ID string) (*Chat, error) {
	u1ID, _ := primitive.ObjectIDFromHex(user1ID)
	u2ID, _ := primitive.ObjectIDFromHex(user2ID)

	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": u1ID, "user2_id": u2ID},
			{"user1_id": u2ID, "user2_id": u1ID},
		},
	}

	var chat Chat
	err := r.db.Collection("chats").FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}
