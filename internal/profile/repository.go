package profile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	UpsertProfile(ctx context.Context, profile *Profile) error
	FindByUserID(ctx context.Context, userID string) (*Profile, error)
	FindByID(ctx context.Context, id string) (*Profile, error)
}

type profileRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &profileRepository{
		collection: db.Collection("profiles"),
	}
}

func (r *profileRepository) UpsertProfile(ctx context.Context, profile *Profile) error {
	filter := bson.M{"user_id": profile.UserID}
	update := bson.M{"$set": profile}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *profileRepository) FindByUserID(ctx context.Context, userID string) (*Profile, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var profile Profile
	err = r.collection.FindOne(ctx, bson.M{"user_id": uID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) FindByID(ctx context.Context, id string) (*Profile, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var profile Profile
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}
