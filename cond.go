package db

import (
	"strings"
)

const (
	OperatorEq       = "="
	OperatorNotEq    = "!="
	OperatorPrefix   = "*="
	OperatorSuffix   = "=*"
	OperatorContains = "*"
	OperatorGt       = ">"
	OperatorGte      = ">="
	OperatorLt       = "<"
	OperatorLte      = "<="
	OperatorRegExp   = "~="
	OperatorIn       = "$in"
	OperatorNotIn    = "$nin"
	OperatorExists   = "$exists"
	OperatorAnd      = "$and"
	OperatorOr       = "$or"
)

type Condition struct {
	Key      string
	Operator string
	Value    interface{}
	Children []Condition
}

type Cond map[string]interface{}

func (c Cond) Conditions() (conditions []Condition) {
	for k, v := range c {
		s := strings.Split(k, " ")
		var (
			key string
			op  string
		)
		if len(s) > 0 {
			key = s[0]
		}
		if len(s) > 1 {
			op = s[1]
		} else {
			op = OperatorEq
		}
		conditions = append(conditions, Condition{
			Key:      key,
			Operator: op,
			Value:    v,
		})
	}
	return
}

type Union struct {
	Operator string
	Filters  []interface{} // 元素可能为 Cond 或 Union
}

func And(ops ...interface{}) interface{} {
	return &Union{
		Operator: OperatorAnd,
		Filters:  ops,
	}
}

func Or(ops ...interface{}) interface{} {
	return &Union{
		Operator: OperatorOr,
		Filters:  ops,
	}
}
