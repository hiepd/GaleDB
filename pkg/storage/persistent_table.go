package storage

import (
	"github.com/hiepd/galedb/pkg/entity"
	"github.com/hiepd/galedb/pkg/index"
)

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

func (pt *PersistentTable) AddRow(row entity.Row) error {
	key, err := pt.Indexes[0].Add(row)
	if err != nil {
		return err
	}
	row.Key = key
	for i := 1; i < len(pt.Indexes); i++ {
		// TODO: Need to rollback everything if one index failed
		if _, err := pt.Indexes[i].Add(row); err != nil {
			return err
		}
	}
	return nil
}
