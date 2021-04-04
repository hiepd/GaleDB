package index

import (
	"container/list"
	"errors"

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

func (si *ScanIndex) Add(row entity.Row) (entity.Key, error) {
	var key entity.Key
	if si.free.Len() == 0 {
		key = entity.Key(len(si.rows) + 1)
		row.Key = key
		si.rows = append(si.rows, &row)
	} else {
		e := si.free.Front()
		freePosition := e.Value.(int)
		key = entity.Key(freePosition + 1)
		row.Key = key
		si.rows[freePosition] = &row
		si.free.Remove(e)
	}
	return key, nil
}

func (si *ScanIndex) Remove(key entity.Key) error {
	position := int(key - 1)
	if position < 0 || position >= len(si.rows) {
		return errors.New("invalid key")
	}
	si.rows[position] = nil
	si.free.PushBack(position)
	return nil
}

func (si *ScanIndex) Get(key entity.Key) (entity.Row, error) {
	position := int(key - 1)
	if position < 0 || position >= len(si.rows) {
		return entity.Row{}, errors.New("invalid key")
	}
	return *si.rows[key-1], nil
}

func (si *ScanIndex) Iterator() Iterator {
	return &ScanIterator{
		index:    si,
		position: -1,
	}
}

func (si *ScanIndex) Size() int {
	return len(si.rows) - si.free.Len()
}

type ScanIterator struct {
	index    *ScanIndex
	position int
}

func (si *ScanIterator) Next() (entity.Row, error) {
	si.position++
	for si.position < len(si.index.rows) && si.index.rows[si.position] == nil {
		si.position++
	}
	if si.position >= len(si.index.rows) {
		return entity.Row{}, EndOfIterator
	}
	return *si.index.rows[si.position], nil
}
