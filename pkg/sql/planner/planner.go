package planner

import (
	"errors"
	"fmt"

	"github.com/hiepd/galedb/pkg/entity"
	"github.com/hiepd/galedb/pkg/index"
	"github.com/hiepd/galedb/pkg/sql/parser"
	"github.com/hiepd/galedb/pkg/storage"
)

var relMap = map[string]Relation{
	"=": Equal,
}

type (
	Node interface {
		Iter() index.Iterator
		Columns() []entity.Column
		Prepare() error
	}

	PlanNode struct {
		Alias string
		Child Node
	}

	Select struct {
		Conditions []*Condition
		PlanNode
	}

	Projection struct {
		Attributes []string
		Cols       []entity.Column
		PlanNode
	}

	Table struct {
		Ref     storage.Table
		RefIter index.Iterator
		PlanNode
	}

	PlanIter struct {
		ChildIter index.Iterator
	}

	SelectIter struct {
		cols   []entity.Column
		conds  []*Condition
		colMap map[string]int
		PlanIter
	}

	ProjectionIter struct {
		colNum   int
		colOrder map[int]int
		PlanIter
	}
)

// Projection Expression
func (proj *Projection) Iter() index.Iterator {
	childColOrder := make(map[string]int)
	childCols := proj.Child.Columns()
	for i, col := range childCols {
		childColOrder[col.Name] = i
	}
	colOrder := make(map[int]int)
	for i, col := range proj.Cols {
		source, ok := childColOrder[col.Name]
		// TODO: handle this better
		if !ok {
			panic("col doesn't exist")
		}
		colOrder[source] = i
	}
	return &ProjectionIter{
		colOrder: colOrder,
		colNum:   len(proj.Cols),
		PlanIter: PlanIter{
			ChildIter: proj.Child.Iter(),
		},
	}
}
func (proj *Projection) Columns() []entity.Column {
	return proj.Cols
}
func (proj *Projection) Prepare() error {
	return proj.resolveColumns()
}
func (proj *Projection) resolveColumns() error {
	if proj.Child == nil {
		return errors.New("no child node")
	}
	colMap := make(map[string]entity.Column)
	childCols := proj.Child.Columns()
	for _, col := range childCols {
		colMap[col.Name] = col
	}
	cols := make([]entity.Column, len(proj.Attributes))
	for i, colName := range proj.Attributes {
		col, ok := colMap[colName]
		if !ok {
			return fmt.Errorf("invalid column %s", colName)
		}
		cols[i] = col
	}
	proj.Cols = cols
	return nil
}

// Select Expression
func (sel *Select) Iter() index.Iterator {
	childCols := sel.Child.Columns()
	colMap := make(map[string]int)
	for i, col := range childCols {
		colMap[col.Name] = i
	}
	return &SelectIter{
		cols:   childCols,
		colMap: colMap,
		conds:  sel.Conditions,
		PlanIter: PlanIter{
			ChildIter: sel.Child.Iter(),
		},
	}
}
func (sel *Select) Columns() []entity.Column {
	return sel.Child.Columns()
}
func (sel *Select) Prepare() error {
	return nil
}

// Table Expression
func (tb *Table) Iter() index.Iterator {
	return tb.RefIter
}
func (tb *Table) Columns() []entity.Column {
	return tb.Ref.Columns()
}
func (tb *Table) Prepare() error {
	switch ref := tb.Ref.(type) {
	case *storage.PersistentTable:
		tb.RefIter = ref.Indexes[0].Iterator()
	default:
		return errors.New("table is not persistent")
	}
	return nil
}

func (iter *SelectIter) Next() (entity.Row, error) {
	for {
		row, err := iter.ChildIter.Next()
		if err != nil {
			return entity.Row{}, err
		}
		if Eval(iter.conds, row, iter.colMap, iter.cols) {
			return row, nil
		}
	}
}

func (iter *ProjectionIter) Next() (entity.Row, error) {
	for {
		row, err := iter.ChildIter.Next()
		if err != nil {
			return entity.Row{}, err
		}
		return iter.project(row, iter.colNum), nil
	}
}

func (iter *ProjectionIter) project(row entity.Row, n int) entity.Row {
	vals := make([]entity.Value, n)
	for k, v := range iter.colOrder {
		vals[v] = row.Values[k]
	}
	return entity.Row{
		Values: vals,
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
	gChild, err := p.parseFromStatement(sel.From)
	if err != nil {
		return nil, err
	}
	if sel.Where != nil {
		child, err := p.parseWhereStatement(sel.Where)
		if err != nil {
			return nil, err
		}
		child.Child = gChild
		return &Projection{
			Attributes: sel.Cols,
			PlanNode: PlanNode{
				Child: child,
			},
		}, nil
	}
	return &Projection{
		Attributes: sel.Cols,
		PlanNode: PlanNode{
			Child: gChild,
		},
	}, nil
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

func (p *Planner) parseWhereStatement(where *parser.Where) (*Select, error) {
	conds := make([]*Condition, len(where.Conditions))
	for i, pcond := range where.Conditions {
		rel, ok := relMap[pcond.Relation]
		if !ok {
			return nil, fmt.Errorf("invalid relation %s", pcond.Relation)
		}
		conds[i] = &Condition{
			Relation: rel,
			LHS:      pcond.LHS,
			RHS:      pcond.RHS,
		}
	}
	return &Select{
		Conditions: conds,
	}, nil
}

func (plan *QueryPlan) Iter() index.Iterator {
	return plan.Root.Iter()
}

func (plan *QueryPlan) Columns() []string {
	cols := plan.Root.Columns()
	res := make([]string, len(cols))
	for i, col := range cols {
		res[i] = col.Name
	}
	return res
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
	case *Select:
		if err := plan.prepare(n.Child); err != nil {
			return err
		}
	default:
	}
	if err := node.Prepare(); err != nil {
		return err
	}
	return nil
}
