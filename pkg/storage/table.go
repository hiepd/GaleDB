package storage

import "github.com/hiepd/galedb/pkg/entity"

type Table interface {
	IsPersistent() bool
	AddRow(row entity.Row) error
}
