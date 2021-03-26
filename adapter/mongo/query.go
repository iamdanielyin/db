package mongo

import (
	"context"
	"github.com/yuyitech/db/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type query struct {
	db   *mongo.Database
	cmd  string
	args []interface{}
}

func (q *query) Iterator() (db.Iterator, error) {
	ctx := context.Background()
	cur, err := q.db.RunCommandCursor(ctx, q.cmd)
	if err != nil {
		return nil, err
	}
	return &iterator{
		ctx: ctx,
		err: err,
		cur: cur,
	}, err
}

func (q *query) One(ptrToStruct interface{}) error {
	sr := q.db.RunCommand(context.Background(), q.cmd)
	if err := sr.Err(); err != nil {
		return err
	}
	return sr.Decode(ptrToStruct)
}

func (q *query) All(sliceOfStruct interface{}) error {
	cur, err := q.db.RunCommandCursor(context.Background(), q.cmd)
	if err != nil {
		return err
	}
	return cur.All(context.Background(), sliceOfStruct)
}
