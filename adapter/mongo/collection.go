package mongo

import (
	"context"
	"github.com/yuyitech/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type mongoCollection struct {
	client *mongoClient
	sess   *db.Connection
	meta   db.Metadata
	db     *mongo.Database
	coll   *mongo.Collection
}

func (c *mongoCollection) Name() string {
	return c.meta.Name
}

func (c *mongoCollection) Metadata() db.Metadata {
	return c.meta
}

func (c *mongoCollection) Session() *db.Connection {
	return c.sess
}

func (c *mongoCollection) InsertOne(v interface{}) (db.InsertOneResult, error) {
	res, err := c.coll.InsertOne(context.Background(), v)
	if err != nil {
		return nil, err
	}
	result := &insertOneResult{result: res}
	return result, nil
}

func (c *mongoCollection) InsertMany(v interface{}) (db.InsertManyResult, error) {
	reflectValue := reflect.ValueOf(v)
	var docs []interface{}
	switch reflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflectValue.Len(); i++ {
			doc := reflectValue.Index(i).Interface()
			docs = append(docs, doc)
		}
	}
	res, err := c.coll.InsertMany(context.Background(), docs)
	if err != nil {
		return nil, err
	}
	result := &insertManyResult{result: res}
	return result, nil
}

func (c *mongoCollection) Find(i ...interface{}) db.Result {
	return &mongoResult{mc: c, conditions: i}
}

type insertOneResult struct {
	result *mongo.InsertOneResult
}

func (i *insertOneResult) StringID() (v string) {
	if i.result != nil {
		id := i.result.InsertedID
		v = objectIdToHex(id)
	}
	return
}

func (i *insertOneResult) IntID() (v int) {
	return
}

type insertManyResult struct {
	result *mongo.InsertManyResult
}

func (i *insertManyResult) StringIDs() (v []string) {
	if i.result != nil {
		for _, item := range i.result.InsertedIDs {
			id := objectIdToHex(item)
			v = append(v, id)
		}
	}
	return
}

func (i *insertManyResult) IntIDs() (v []int) {
	return
}

func objectIdToHex(id interface{}) (v string) {
	switch id.(type) {
	case primitive.ObjectID:
		v = id.(primitive.ObjectID).Hex()
	case string:
		v = id.(string)
	}
	return
}
