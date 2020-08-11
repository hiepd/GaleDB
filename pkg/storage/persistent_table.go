package storage

import "github.com/hiepd/galedb/pkg/index"

type PersistentTable struct {
	Columns []*Column
	Indexes []index.Index
}

func NewPersisentTable() Table {
	indexes := []index.Index{index.NewScanIndex()}
	return &PersistentTable{
		Indexes: indexes,
	}
}

func (pt *PersistentTable) IsPersistent() bool {
	return true
}
