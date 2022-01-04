package mongo

import (
	"fmt"
	"github.com/iamdanielyin/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

func QueryFilter(meta db.Metadata, filters ...interface{}) (d bson.D) {
	execCond := func(v *db.Cond) {
		for _, item := range v.Entries() {
			condition := parseCondition(meta, &item)
			if condition != nil {
				d = append(d, *condition)
			}
		}
	}
	execUnion := func(v *db.Union) {
		var arr bson.A
		for _, each := range v.Conditions() {
			eachD := QueryFilter(meta, each)
			if len(eachD) > 0 {
				arr = append(arr, eachD)
			}
		}
		if v.Operator() == db.OperatorOr {
			d = append(d, bson.E{Key: "$or", Value: arr})
		} else {
			d = append(d, bson.E{Key: "$and", Value: arr})
		}
	}
	for _, filter := range filters {
		switch v := filter.(type) {
		case db.Cond:
			execCond(&v)
		case *db.Cond:
			execCond(v)
		case db.Union:
			execUnion(&v)
		case *db.Union:
			execUnion(v)
		}
	}
	return
}

func parseCondition(meta db.Metadata, item *db.ConditionEntry) *bson.E {
	key := item.Key

	if f, has := meta.FieldByName(item.Key); has {
		key = f.MustNativeName()
	}

	switch item.Operator {
	case db.OperatorEq:
		return &bson.E{Key: key, Value: item.Value}
	case db.OperatorNotEq:
		return &bson.E{Key: key, Value: bson.D{{Key: "$ne", Value: item.Value}}}
	case db.OperatorPrefix:
		return &bson.E{Key: key, Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", item.Value), Options: "gim"}}
	case db.OperatorSuffix:
		return &bson.E{Key: key, Value: primitive.Regex{Pattern: fmt.Sprintf("%v$", item.Value), Options: "gim"}}
	case db.OperatorContains:
		return &bson.E{Key: key, Value: primitive.Regex{Pattern: fmt.Sprintf("%v", item.Value), Options: "gim"}}
	case db.OperatorGt:
		return &bson.E{Key: key, Value: bson.D{{Key: "$gt", Value: item.Value}}}
	case db.OperatorGte:
		return &bson.E{Key: key, Value: bson.D{{Key: "$gte", Value: item.Value}}}
	case db.OperatorLt:
		return &bson.E{Key: key, Value: bson.D{{Key: "$lt", Value: item.Value}}}
	case db.OperatorLte:
		return &bson.E{Key: key, Value: bson.D{{Key: "$lte", Value: item.Value}}}
	case db.OperatorRegExp:
		var (
			s       = item.Value.(string)
			pattern string
			options string
		)
		if lastIdx := strings.LastIndex(s, "/"); strings.HasPrefix(s, "/") && lastIdx > 0 {
			pattern = s[1 : lastIdx+1]
			options = s[lastIdx+1 : len(s)-1]
		} else {
			pattern = s
		}
		return &bson.E{Key: key, Value: primitive.Regex{Pattern: pattern, Options: options}}

	case db.OperatorIn:
		return &bson.E{Key: key, Value: bson.D{{Key: "$in", Value: item.Value}}}
	case db.OperatorNotIn:
		return &bson.E{Key: key, Value: bson.D{{Key: "$nin", Value: item.Value}}}
	case db.OperatorExists:
		return &bson.E{Key: key, Value: bson.D{{Key: "$exists", Value: item.Value}}}
	}
	return nil
}
