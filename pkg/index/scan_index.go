package index

import (
	"container/list"

	"github.com/hiepd/galedb/pkg/entity"
)

type ScanIndex struct {
	rows []*entity.Row
	free *list.List
}

func NewScanIndex() Index {
	return &ScanIndex{
		rows: make([]*entity.Row, 0),
		free: list.New(),
	}
}

func (si *ScanIndex) Add(row entity.Row) error {
	if si.free.Len() == 0 {
		key := entity.Key(len(si.rows) + 1)
		row.Key = key
		si.rows = append(si.rows, &row)
	} else {
		e := si.free.Front()
		freePosition := e.Value.(int)
		key := entity.Key(freePosition + 1)
		row.Key = key
		si.rows[freePosition] = &row
		si.free.Remove(e)
	}
	return nil
}

func (si *ScanIndex) Remove(key entity.Key) error {
	position := key - 1
	si.rows[position] = nil
	si.free.PushBack(position)
	return nil
}

func (si *ScanIndex) Get(key entity.Key) (entity.Row, error) {
	return *si.rows[key-1], nil
}

func (si *ScanIndex) Iterator() Iterator {
	return &ScanIterator{
		index:    si,
		position: -1,
	}
}

func (si *ScanIndex) Size() int {
	return len(si.rows)
}

type ScanIterator struct {
	index    *ScanIndex
	position int
}

func (si *ScanIterator) Next() bool {
	si.position++
	return si.position < si.index.Size()
}

func (si *ScanIterator) Current() entity.Row {
	return *si.index.rows[si.position]
}
