package profile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name            string             `bson:"name" json:"name" validate:"required"`
	Age             int                `bson:"age" json:"age" validate:"required,min=18,max=100"`
	Gender          string             `bson:"gender" json:"gender" validate:"required,oneof=male female"`
	Church          string             `bson:"church" json:"church" validate:"required"`
	Location        string             `bson:"location" json:"location" validate:"required"`
	Values          []string           `bson:"values" json:"values"`
	PreferredGender string             `bson:"preferred_gender" json:"preferred_gender"`
	MinAge          int                `bson:"min_age" json:"min_age"`
	MaxAge          int                `bson:"max_age" json:"max_age"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type PublicProfile struct {
	Name     string   `json:"name"`
	Age      int      `json:"age"`
	Gender   string   `json:"gender"`
	Church   string   `json:"church"`
	Location string   `json:"location"`
	Values   []string `json:"values"`
}
