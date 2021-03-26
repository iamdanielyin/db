package mongo

import (
	"github.com/yuyitech/db/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestCond(t *testing.T) {
	var bs = new(bson.D)
	bs = ParseFilter(nil, db.Cond{
		"a":    "x",
		"b >":  10,
		"c <":  10,
		"d >=": 10,
		"e <=": 10,
		"f !=": 10,
		"g *~": "x",
		"h ~*": "x",
		"i *":  "x",
		"j in": []interface{}{
			1, 2, 3,
		},
		"k nin": []interface{}{
			1, 2, 3,
		},
		"l ||": db.D{
			">": 10,
			"<": 100,
		},
		"m &&": db.D{
			">": 10,
			"<": 100,
		},
	})
	t.Log(bs)

	bs = ParseFilter(nil, db.Or(
		db.Cond{
			"a":   "x",
			"b >": 10,
		},
		db.Cond{
			"a":   "x",
			"b <": 100,
		},
	))
	t.Log(bs)

	bs = ParseFilter(nil, db.And(
		db.Cond{
			"a":   "x",
			"b >": 10,
		},
		db.Cond{
			"a":   "x",
			"b <": 100,
		},
	))
	t.Log(bs)
}
