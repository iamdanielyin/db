package db

import (
	"fmt"
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

type Conditional interface {
	Operator() string
	Conditions() []Conditional
}

type ConditionEntry struct {
	Key      string
	Operator string
	Value    interface{}
	Children []ConditionEntry
}

type Cond map[string]interface{}

func NewCond() Cond {
	return Cond{}
}

func (c Cond) Op(key, operator string, value interface{}) Cond {
	key = strings.TrimSpace(key)
	operator = strings.TrimSpace(operator)
	if key != "" {
		c[fmt.Sprintf("%s %s", key, operator)] = value
	}
	return c
}

func (c Cond) Eq(key string, value interface{}) Cond {
	return c.Op(key, OperatorEq, value)
}

func (c Cond) NotEq(key string, value interface{}) Cond {
	return c.Op(key, OperatorNotEq, value)
}

func (c Cond) Prefix(key string, value interface{}) Cond {
	return c.Op(key, OperatorPrefix, value)
}

func (c Cond) Suffix(key string, value interface{}) Cond {
	return c.Op(key, OperatorSuffix, value)
}

func (c Cond) Contains(key string, value interface{}) Cond {
	return c.Op(key, OperatorContains, value)
}

func (c Cond) Gt(key string, value interface{}) Cond {
	return c.Op(key, OperatorGt, value)
}

func (c Cond) Gte(key string, value interface{}) Cond {
	return c.Op(key, OperatorGte, value)
}

func (c Cond) Lt(key string, value interface{}) Cond {
	return c.Op(key, OperatorLt, value)
}

func (c Cond) Lte(key string, value interface{}) Cond {
	return c.Op(key, OperatorLte, value)
}

func (c Cond) RegExp(key string, value interface{}) Cond {
	return c.Op(key, OperatorRegExp, value)
}

func (c Cond) In(key string, value interface{}) Cond {
	return c.Op(key, OperatorIn, value)
}

func (c Cond) NotIn(key string, value interface{}) Cond {
	return c.Op(key, OperatorNotIn, value)
}

func (c Cond) Exists(key string, value interface{}) Cond {
	return c.Op(key, OperatorExists, value)
}

func (c Cond) Entries() (entries []ConditionEntry) {
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
		entries = append(entries, ConditionEntry{
			Key:      key,
			Operator: op,
			Value:    v,
		})
	}
	return
}

func (c Cond) Operator() string {
	return OperatorAnd
}

func (c Cond) Conditions() []Conditional {
	return []Conditional{c}
}

type Union struct {
	operator   string
	conditions []Conditional // 元素可能为 Cond 或 Union
}

func (u *Union) Operator() string {
	return u.operator
}

func (u *Union) Conditions() []Conditional {
	return u.conditions
}

func NewUnion(op string, v []Conditional) Conditional {
	op = strings.TrimSpace(op)
	if op != "" && len(v) > 0 {
		var conditions []Conditional
		for _, item := range v {
			if item != nil && len(item.Conditions()) > 0 {
				conditions = append(conditions, item)
			}
		}
		if len(conditions) > 0 {
			return &Union{
				operator:   op,
				conditions: conditions,
			}
		}
	}
	return nil
}

func And(v ...Conditional) Conditional {
	return NewUnion(OperatorAnd, v)
}

func Or(v ...Conditional) Conditional {
	return NewUnion(OperatorAnd, v)
}
