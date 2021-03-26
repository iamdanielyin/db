package db

import (
	"github.com/yuyitech/db/internal/json"
	"github.com/yuyitech/db/internal/reflectx"
	"strings"
)

const (
	OperatorOr       = "||"
	OperatorAnd      = "&&"
	OperatorEq       = "="
	OperatorGt       = ">"
	OperatorGte      = ">="
	OperatorLt       = "<"
	OperatorLte      = "<="
	OperatorNotEq    = "!="
	OperatorPrefix   = "~*"
	OperatorSuffix   = "*~"
	OperatorContains = "*"
	OperatorIn       = "in"
	OperatorNotIn    = "nin"
)

type Cond map[string]interface{}

type Comparison struct {
	SubComparisons []Comparison
	Key            string
	Operator       string
	Value          interface{}
}

func (c *Cond) Comparisons(frame ...func(*Comparison)) (cs []Comparison) {
	for k, v := range *c {
		s := strings.Split(k, " ")
		var cmp Comparison
		if len(s) > 0 {
			cmp.Key = s[0]
		}
		if len(s) > 1 {
			cmp.Operator = s[1]
		} else {
			cmp.Operator = OperatorEq
		}

		switch cmp.Operator {
		case OperatorAnd, OperatorOr:
			var m map[string]interface{}
			if sv, ok := v.(map[string]interface{}); ok {
				m = sv
			} else {
				_ = json.Copy(v, &m)
			}
			for sk, sv := range m {
				cmp.SubComparisons = append(cmp.SubComparisons, Comparison{
					Key:      cmp.Key,
					Operator: sk,
					Value:    sv,
				})
			}
		case OperatorIn, OperatorNotIn:
			cmp.Value = reflectx.ToInterfaceArray(v)
		default:
			cmp.Value = v
		}
		if len(frame) > 0 && frame[0] != nil {
			frame[0](&cmp)
		}
		cs = append(cs, cmp)
	}
	return
}

type Union struct {
	Filters  []Cond
	Operator string
}

func Or(filters ...Cond) *Union {
	return &Union{
		Filters:  filters,
		Operator: OperatorOr,
	}
}

func And(filters ...Cond) *Union {
	return &Union{
		Filters:  filters,
		Operator: OperatorAnd,
	}
}
