package parser

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type (
	Statement interface {
		iStatement()
		String() string
	}

	Select struct {
		Cols  []string
		From  *From
		Where *Where
	}

	From struct {
		TableName string
	}

	Where struct {
		Conditions []*Condition
	}

	Condition struct {
		Relation string
		LHS      interface{}
		RHS      interface{}
	}
)

func (*Select) iStatement() {}
func (sel *Select) String() string {
	return fmt.Sprintf("SELECT %s\n--%s", sel.Cols, sel.From.String())
}

func (*From) iStatement() {}
func (from *From) String() string {
	return fmt.Sprintf("FROM %s", from.TableName)
}

func (*Condition) iStatement() {}
func (cond *Condition) String() string {
	return fmt.Sprintf("%v %s %v", cond.LHS, cond.Relation, cond.RHS)
}

func (*Where) iStatement() {}
func (where *Where) String() string {
	conds := make([]string, len(where.Conditions))
	for i, cond := range where.Conditions {
		conds[i] = cond.String()
	}
	return fmt.Sprintf("WHERE %s", strings.Join(conds, "AND"))
}

func NewSelect(cols []string, from Statement, where *Where) Statement {
	logrus.Infof("colexpr: %s", cols)
	return &Select{
		Cols:  cols,
		From:  from.(*From),
		Where: where,
	}
}

func NewFrom(tableName string) Statement {
	return &From{
		TableName: tableName,
	}
}

func NewCondition(relation string, lhs interface{}, rhs interface{}) *Condition {
	return &Condition{
		Relation: relation,
		LHS:      lhs,
		RHS:      rhs,
	}
}

func NewWhere(conds []*Condition) *Where {
	return &Where{
		Conditions: conds,
	}
}
