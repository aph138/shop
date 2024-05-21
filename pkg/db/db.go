package db

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// in millisecond
const TIMEOUT = 1500

// nil opt = default opt
func NewDB(address string, logger *slog.Logger, opt *options.ClientOptions) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*TIMEOUT)
	defer cancel()
	if opt == nil {
		opt = options.Client().ApplyURI(address)
	}
	dbClient, err := mongo.Connect(ctx, opt)
	if err != nil {
		return nil, err
	}
	if err = dbClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}
	return dbClient, nil
}
