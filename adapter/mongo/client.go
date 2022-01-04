package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"time"
)

type mongoClient struct {
	adapter *mongoAdapter
	source  db.DataSource
	cs      connstring.ConnString
	client  *mongo.Client
	logger  db.Logger
}

func (c *mongoClient) StartTransaction() (db.Tx, error) {
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := c.client.StartSession(opts)
	if err != nil {
		return nil, db.Errorf(`%v`, err)
	}
	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	if err := sess.StartTransaction(txnOpts); err != nil {
		return nil, db.Errorf(`%v`, err)
	}
	mt := &mongoTx{
		ctx:       context.Background(),
		client:    c,
		mongoSess: sess,
	}
	return mt, nil
}

func (c *mongoClient) WithTransaction(fn func(db.Tx) error) error {
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := c.client.StartSession(opts)
	if err != nil {
		return db.Errorf(`%v`, err)
	}
	defer sess.EndSession(context.Background())
	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	_, err = sess.WithTransaction(context.Background(), func(sessCtx mongo.SessionContext) (interface{}, error) {
		mt := &mongoTx{
			ctx:       sessCtx,
			client:    c,
			mongoSess: sess,
		}
		err := fn(mt)
		return nil, err
	}, txnOpts)
	return err
}

func (c *mongoClient) Name() string {
	return c.source.Name
}

func (c *mongoClient) Logger() db.Logger {
	return c.logger
}

func (c *mongoClient) Source() db.DataSource {
	return c.source
}

func (c *mongoClient) Model(metadata db.Metadata) db.Collection {
	mdb := c.client.Database(c.cs.Database)
	coll := mdb.Collection(metadata.MustNativeName())
	return &mongoCollection{
		client: c,
		sess:   metadata.Session(),
		meta:   metadata,
		db:     mdb,
		coll:   coll,
	}
}

func (c *mongoClient) Disconnect(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}
