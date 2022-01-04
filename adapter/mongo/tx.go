package mongo

import (
	"context"
	"github.com/iamdanielyin/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoTx struct {
	ctx       context.Context
	client    *mongoClient
	mongoSess mongo.Session
}

func (mt *mongoTx) Model(name string) db.Collection {
	meta, err := db.LookupMetadata(name)
	if err != nil {
		panic(err)
	}
	return mt.client.Model(meta)
}

func (mt *mongoTx) Commit() error {
	defer mt.close()
	if err := mt.mongoSess.CommitTransaction(mt.ctx); err != nil {
		return db.Errorf(`%v`, err)
	}
	return nil
}

func (mt *mongoTx) Rollback() error {
	defer mt.close()
	if err := mt.mongoSess.AbortTransaction(mt.ctx); err != nil {
		return db.Errorf(`%v`, err)
	}
	return nil
}

func (mt *mongoTx) close() {
	mt.mongoSess.EndSession(mt.ctx)
}
