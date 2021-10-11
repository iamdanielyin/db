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
}

func init() {
	db.RegisterAdapter(Adapter, &mongoAdapter{})
}

func (a *mongoAdapter) Name() string {
	return Adapter
}

func (a *mongoAdapter) Connect(source db.DataSource) (db.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(source.URI))
	if err != nil {
		return nil, err
	}
	client := &mongoClient{
		adapter: a,
		source:  source,
		client:  c,
	}
	return client, nil
}
