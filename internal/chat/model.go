package chat

import (
	"church-match-api/internal/profile"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User1ID   primitive.ObjectID `bson:"user1_id" json:"user1_id"`
	User2ID   primitive.ObjectID `bson:"user2_id" json:"user2_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatID    primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Content   string             `bson:"content" json:"content" validate:"required,max=2000"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type ChatSummary struct {
	Chat        Chat                 `json:"chat"`
	OtherUser   profile.PublicProfile `json:"other_user"`
	LastMessage *Message             `json:"last_message,omitempty"`
}

type MessageWithSender struct {
	Message `bson:",inline"`
	Sender  profile.PublicProfile `json:"sender"`
}

type WSMessage struct {
	Type    string      `json:"type"` // "new_message"
	ChatID  string      `json:"chat_id"`
	Message Message     `json:"message"`
}
