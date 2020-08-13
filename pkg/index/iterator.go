package index

import "github.com/hiepd/galedb/pkg/entity"

type Iterator interface {
	Next() bool
	Current() (entity.Row, error)
}
