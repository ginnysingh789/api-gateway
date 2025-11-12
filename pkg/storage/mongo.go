package storage

import (
	"context"
	"time"

	"api-gateway/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	*mongo.Client
	Database *mongo.Database
}

func NewMongoClient(cfg config.MongoDBConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoClient{
		Client:   client,
		Database: client.Database(cfg.Database),
	}, nil
}

func (m *MongoClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}
