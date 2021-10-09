package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const Adapter = `mongo`

type mongoAdapter struct {
	source db.DataSource
	client *mongo.Client
}

func init() {
	db.RegisterAdapter(Adapter, &mongoAdapter{})
}

func (a *mongoAdapter) Connect(source db.DataSource) (db.IConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(source.URI))
	if err != nil {
		return nil, err
	}
	a.source = source
	a.client = client
	return &mongoConnection{adapter: a}, nil
}

func (a *mongoAdapter) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if a.client != nil {
		return a.client.Disconnect(ctx)
	}
	return nil
}
