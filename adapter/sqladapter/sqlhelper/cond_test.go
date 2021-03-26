package sqlhelper

import (
	"github.com/yuyitech/db/pkg/db"
	"testing"
)

func TestParseFilter(t *testing.T) {
	var cmp *Compound
	cmp = ParseFilter(db.Cond{
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
	t.Log(cmp.CombineStmts())

	cmp = ParseFilter(db.Or(
		db.Cond{
			"a":   "x",
			"b >": 10,
		},
		db.Cond{
			"a":   "x",
			"b <": 100,
		},
	))
	t.Log(cmp.CombineStmts())

	cmp = ParseFilter(db.And(
		db.Cond{
			"a":   "x",
			"b >": 10,
		},
		db.Cond{
			"a":   "x",
			"b <": 100,
		},
	))
	t.Log(cmp.CombineStmts())
}
