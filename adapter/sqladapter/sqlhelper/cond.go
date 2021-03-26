package sqlhelper

import (
	"bytes"
	"fmt"
	"github.com/yuyitech/db/pkg/db"
	"strings"
)

type Compound struct {
	Op    string
	Stmts []string
	Args  []interface{}
}

func (c *Compound) CombineStmts() string {
	var buffer bytes.Buffer
	for index, stmt := range c.Stmts {
		buffer.WriteString(fmt.Sprintf("(%s)", stmt))
		if index < len(c.Stmts)-1 {
			buffer.WriteString(fmt.Sprintf(" %s ", c.Op))
		}
	}
	return buffer.String()
}

func ParseFilter(a interface{}, frame ...func(*db.Comparison)) (cmp *Compound) {
	cmp = &Compound{}
	switch v := a.(type) {
	case db.Cond:
		cmp.Op = "AND"
		for _, item := range v.Comparisons(frame...) {
			parseComparison(&item, cmp)
		}
	case *db.Union:
		cmp.Op = map[string]string{
			"||": "OR",
			"&&": "AND",
		}[v.Operator]
		for _, filter := range v.Filters {
			sub := ParseFilter(filter)
			cmp.Stmts = append(cmp.Stmts, sub.CombineStmts())
			cmp.Args = append(cmp.Args, sub.Args...)
		}
	}
	return cmp
}

func parseComparison(item *db.Comparison, cmp *Compound) {
	switch item.Operator {
	case db.OperatorEq:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s = ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorGt:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s > ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorGte:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s >= ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorLt:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s < ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorLte:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s <= ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorNotEq:
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s <> ?", item.Key))
		cmp.Args = append(cmp.Args, item.Value)
	case db.OperatorPrefix:
		if s, ok := item.Value.(string); ok {
			cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s LIKE ?", item.Key))
			cmp.Args = append(cmp.Args, fmt.Sprintf("%s%s", s, "%"))
		}
	case db.OperatorSuffix:
		if s, ok := item.Value.(string); ok {
			cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s LIKE ?", item.Key))
			cmp.Args = append(cmp.Args, fmt.Sprintf("%s%s", "%", s))
		}
	case db.OperatorContains:
		if s, ok := item.Value.(string); ok {
			cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s LIKE ?", item.Key))
			cmp.Args = append(cmp.Args, fmt.Sprintf("%s%s%s", "%", s, "%"))
		}
	case db.OperatorIn, db.OperatorNotIn:
		var op string
		if item.Operator == db.OperatorIn {
			op = "IN"
		} else {
			op = "NOT IN"
		}
		var (
			placeholder string
			args        []interface{}
		)

		values := item.Value.([]interface{})
		if len(values) < 1 {
			placeholder, args = "(NULL)", []interface{}{}
			break
		}
		placeholder, args = "(?"+strings.Repeat(", ?", len(values)-1)+")", values
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("%s %s %s", item.Key, op, placeholder))
		cmp.Args = append(cmp.Args, args...)
	case db.OperatorOr, db.OperatorAnd:
		var subCmp Compound
		for _, subItem := range item.SubComparisons {
			parseComparison(&db.Comparison{
				Key:      subItem.Key,
				Operator: subItem.Operator,
				Value:    subItem.Value,
			}, &subCmp)
		}
		subCmp.Op = map[string]string{
			"||": "OR",
			"&&": "AND",
		}[item.Operator]
		cmp.Stmts = append(cmp.Stmts, fmt.Sprintf("(%s)", subCmp.CombineStmts()))
		cmp.Args = append(cmp.Args, subCmp.Args...)
	}
}
