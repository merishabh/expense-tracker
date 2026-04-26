package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Memory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	LastUsed  time.Time          `bson:"last_used" json:"last_used"`
}

func (m *MongoClient) SaveMemory(mem Memory) error {
	collection := m.Database.Collection("chat_memories")
	mem.ID = primitive.NewObjectID()
	mem.CreatedAt = time.Now()
	mem.LastUsed = time.Now()
	_, err := collection.InsertOne(m.Ctx, mem)
	if err != nil {
		return fmt.Errorf("failed to save memory: %v", err)
	}
	return nil
}

func (m *MongoClient) GetAllMemories() ([]Memory, error) {
	collection := m.Database.Collection("chat_memories")
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(m.Ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memories: %v", err)
	}
	defer cursor.Close(m.Ctx)

	var memories []Memory
	if err := cursor.All(m.Ctx, &memories); err != nil {
		return nil, fmt.Errorf("failed to decode memories: %v", err)
	}
	return memories, nil
}
