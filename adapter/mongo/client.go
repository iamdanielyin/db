package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type mongoClient struct {
	adapter *mongoAdapter
	source  db.DataSource
	client  *mongo.Client
}

func (c *mongoClient) Name() string {
	return c.source.Name
}

func (c *mongoClient) Source() db.DataSource {
	return c.source
}

func (c *mongoClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}
