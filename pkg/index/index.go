package index

import (
	"github.com/hiepd/galedb/pkg/entity"
)

type Index interface {
	Add(row entity.Row) (entity.Key, error)
	Remove(key entity.Key) error
	Get(key entity.Key) (entity.Row, error)
	Iterator() Iterator
	Size() int
}
