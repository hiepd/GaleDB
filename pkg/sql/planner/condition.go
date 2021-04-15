package planner

import (
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/hiepd/galedb/pkg/entity"
)

const (
	Equal = iota + 1
)

type Relation int

type Condition struct {
	Relation Relation
	LHS      interface{}
	RHS      interface{}
}

func Eval(conds []*Condition, row entity.Row, colMap map[string]int, cols []entity.Column) bool {
	for _, cond := range conds {
		if !cond.Eval(row, colMap, cols) {
			return false
		}
	}
	return true
}

func (c *Condition) Eval(row entity.Row, colMap map[string]int, cols []entity.Column) bool {
	colName, ok := c.LHS.(string)
	if !ok {
		return false
	}
	id, ok := colMap[colName]
	if !ok {
		return false
	}
	col := cols[id]
	var rval interface{}
	switch col.Kind {
	case reflect.String:
		rval, ok = c.RHS.(string)
		if !ok {
			return false
		}
	case reflect.Int:
		rval, ok = c.RHS.(int)
		if !ok {
			return false
		}
	default:
		return false
	}
	return compare(c.Relation, row.Values[id], rval)
}

func compare(relation Relation, a, b interface{}) bool {
	logrus.Debugf("comparing %v %d %v", a, relation, b)
	switch relation {
	case Equal:
		return a == b
	}
	return false
}
