package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"reflect"
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

func (a *mongoAdapter) Connect(parent context.Context, source db.DataSource, logger db.Logger) (db.Client, error) {
	cs, err := connstring.Parse(source.URI)
	if err != nil {
		return nil, err
	}
	var c *mongo.Client
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()

	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			logger.INFO("[db:%d] %v\n",
				evt.RequestID,
				evt.Command,
			)
		},
		Succeeded: func(_ context.Context, evt *event.CommandSucceededEvent) {
			logger.INFO("[db:%d] %v\n",
				evt.RequestID,
				evt.Reply,
			)
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			logger.ERROR("[db:%d] %v\n",
				evt.RequestID,
				evt.Failure,
			)
		},
	}
	var codec bsoncodec.ValueCodec
	if v, err := bsoncodec.NewStructCodec(structTagParser); err != nil {
		return nil, err
	} else {
		codec = v
	}
	registry := bson.NewRegistryBuilder().
		RegisterDefaultEncoder(reflect.Struct, codec).
		RegisterDefaultDecoder(reflect.Struct, codec).
		Build()
	opts := options.Client().ApplyURI(source.URI).SetMonitor(cmdMonitor).SetRegistry(registry)
	c, err = mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	client := &mongoClient{
		adapter: a,
		source:  source,
		cs:      cs,
		client:  c,
		logger:  logger,
	}
	return client, nil
}
