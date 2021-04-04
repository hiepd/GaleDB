package planner

import (
	"errors"

	"github.com/hiepd/galedb/pkg/entity"
	"github.com/hiepd/galedb/pkg/index"
	"github.com/hiepd/galedb/pkg/sql/parser"
	"github.com/hiepd/galedb/pkg/storage"
)

type (
	Node interface {
		iNode()
		Iter() index.Iterator
	}

	PlanNode struct {
		Alias string
		Child Node
	}

	Select struct {
		Predicates []Predicate
		PlanNode
	}

	Projection struct {
		Attributes []string
		PlanNode
	}

	Table struct {
		Ref     storage.Table
		RefIter index.Iterator
		PlanNode
	}

	Predicate struct{}

	PlanIter struct {
		ChildIter index.Iterator
	}

	SelectIter struct {
		PlanIter
	}

	ProjectionIter struct {
		PlanIter
	}
)

func (*Select) iNode()     {}
func (*Projection) iNode() {}
func (*Table) iNode()      {}
func (sel *Select) Iter() index.Iterator {
	return &SelectIter{
		PlanIter: PlanIter{
			ChildIter: sel.Child.Iter(),
		},
	}
}
func (proj *Projection) Iter() index.Iterator {
	return &SelectIter{
		PlanIter: PlanIter{
			ChildIter: proj.Child.Iter(),
		},
	}
}
func (tb *Table) Iter() index.Iterator {
	return tb.RefIter
}

func (iter *SelectIter) Next() (entity.Row, error) {
	for {
		row, err := iter.ChildIter.Next()
		if err != nil {
			return entity.Row{}, err
		}
		return row, nil
	}
}

func (iter *ProjectionIter) Next() (entity.Row, error) {
	for {
		row, err := iter.ChildIter.Next()
		if err != nil {
			return entity.Row{}, err
		}
		return row, nil
	}
}

type Planner struct {
	Database *storage.Database
}

type QueryPlan struct {
	Root Node
}

func New(db *storage.Database) *Planner {
	return &Planner{
		Database: db,
	}
}

func (p *Planner) Prepare(statement parser.Statement) (*QueryPlan, error) {
	switch stmt := statement.(type) {
	case *parser.Select:
		plan, err := p.buildQueryPlan(stmt)
		if err != nil {
			return nil, err
		}
		err = plan.prepare(plan.Root)
		return plan, err
	default:
		return nil, errors.New("unsupported statement")
	}
}

func (p *Planner) buildQueryPlan(sel *parser.Select) (*QueryPlan, error) {
	root, err := p.parseSelectStatement(sel)
	if err != nil {
		return nil, err
	}
	plan := &QueryPlan{
		Root: root,
	}
	return plan, nil
}

func (p *Planner) parseSelectStatement(sel *parser.Select) (Node, error) {
	child, err := p.parseFromStatement(sel.From)
	if err != nil {
		return nil, err
	}
	node := &Projection{
		PlanNode: PlanNode{
			Child: child,
		},
	}
	return node, nil
}

func (p *Planner) parseFromStatement(from *parser.From) (Node, error) {
	table, err := p.Database.GetTable(from.TableName)
	if err != nil {
		return nil, err
	}
	return &Table{
		Ref: table,
	}, nil
}

func (plan *QueryPlan) Iter() index.Iterator {
	return plan.Root.Iter()
}

func (plan *QueryPlan) prepare(node Node) error {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Projection:
		if err := plan.prepare(n.Child); err != nil {
			return err
		}
	case *Table:
		switch tb := n.Ref.(type) {
		case *storage.PersistentTable:
			n.RefIter = tb.Indexes[0].Iterator()
		default:
			return errors.New("table is not persistent")
		}
	}
	return nil
}
