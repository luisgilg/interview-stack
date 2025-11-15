package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/example/interview-stack/go-service/internal/config"
)

// NewMongoCollection returns a collection reference after verifying the connection.
func NewMongoCollection(ctx context.Context, cfg config.MongoConfig) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout.Duration())
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetServerSelectionTimeout(cfg.ConnectTimeout.Duration()).
		SetConnectTimeout(cfg.ConnectTimeout.Duration()).
		SetSocketTimeout(cfg.OperationTimeout.Duration())

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client.Database(cfg.Database).Collection(cfg.Collection), nil
}
