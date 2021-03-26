package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type iterator struct {
	ctx context.Context
	err error
	cur *mongo.Cursor
}

func (i *iterator) Next(ptrToStruct interface{}) bool {
	next := i.cur.Next(i.ctx)
	if next {
		i.err = i.cur.Decode(ptrToStruct)
	}
	return next
}

func (i *iterator) Err() error {
	return i.err
}

func (i *iterator) Close() error {
	if i.cur != nil {
		return i.cur.Close(i.ctx)
	}
	return nil
}
