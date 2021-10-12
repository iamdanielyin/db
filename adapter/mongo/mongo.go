package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
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

func (a *mongoAdapter) Connect(parent context.Context, source db.DataSource) (db.Client, error) {
	cs, err := connstring.Parse(source.URI)
	if err != nil {
		return nil, err
	}
	var c *mongo.Client
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	c, err = mongo.Connect(ctx, options.Client().ApplyURI(source.URI))
	if err != nil {
		return nil, err
	}
	client := &mongoClient{
		adapter: a,
		source:  source,
		cs:      cs,
		client:  c,
	}
	return client, nil
}
