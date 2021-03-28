package mongo

import (
	"fmt"
	"github.com/yuyitech/db/internal/json"
	"github.com/yuyitech/db/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ParseFilter(frame func(*db.Comparison), filters ...interface{}) (bs *bson.D) {
	bs = &bson.D{}
	for _, a := range filters {
		switch v := a.(type) {
		case db.Cond:
			for _, item := range v.Comparisons(frame) {
				parseComparison(&item, bs)
			}
		//case []interface{}:
		//	sbs := ParseFilter(frame, v)
		//	bs = sbs
		case *db.Union:
			var a bson.A
			for _, filter := range v.Filters {
				sub := ParseFilter(frame, filter)
				a = append(a, sub)
			}
			switch v.Operator {
			case db.OperatorOr:
				*bs = append(*bs, bson.E{Key: "$or", Value: a})
			case db.OperatorAnd:
				*bs = append(*bs, bson.E{Key: "$and", Value: a})
			}
		}
	}
	return
}

func parseComparison(item *db.Comparison, bs *bson.D) {
	switch item.Operator {
	case db.OperatorEq:
		*bs = append(*bs, bson.E{Key: item.Key, Value: item.Value})
	case db.OperatorGt:
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$gt", item.Value}}})
	case db.OperatorGte:
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$gte", item.Value}}})
	case db.OperatorLt:
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$lt", item.Value}}})
	case db.OperatorLte:
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$lte", item.Value}}})
	case db.OperatorNotEq:
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$ne", item.Value}}})
	case db.OperatorPrefix:
		if s, ok := item.Value.(string); ok {
			*bs = append(*bs, bson.E{Key: item.Key, Value: primitive.Regex{Pattern: fmt.Sprintf("^%s", s), Options: "im"}})
		}
	case db.OperatorSuffix:
		if s, ok := item.Value.(string); ok {
			*bs = append(*bs, bson.E{Key: item.Key, Value: primitive.Regex{Pattern: fmt.Sprintf("%s$", s), Options: "im"}})
		}
	case db.OperatorContains:
		if s, ok := item.Value.(string); ok {
			*bs = append(*bs, bson.E{Key: item.Key, Value: primitive.Regex{Pattern: fmt.Sprintf("%s", s), Options: "im"}})
		}
	case db.OperatorIn:
		var a bson.A
		_ = json.Copy(item.Value, &a)
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$in", a}}})
	case db.OperatorNotIn:
		var a bson.A
		_ = json.Copy(item.Value, &a)
		*bs = append(*bs, bson.E{Key: item.Key, Value: bson.D{{"$nin", a}}})
	case db.OperatorOr, db.OperatorAnd:
		var a bson.A
		for _, subItem := range item.SubComparisons {
			var subD bson.D
			parseComparison(&db.Comparison{
				Key:      subItem.Key,
				Operator: subItem.Operator,
				Value:    subItem.Value,
			}, &subD)
			a = append(a, subD)
		}
		if item.Operator == db.OperatorOr {
			*bs = append(*bs, bson.E{Key: "$or", Value: a})
		} else {
			*bs = append(*bs, bson.E{Key: "$and", Value: a})
		}
	}
}
