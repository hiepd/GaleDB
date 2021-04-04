package index

import (
	"errors"

	"github.com/hiepd/galedb/pkg/entity"
)

var EndOfIterator = errors.New("end of iterator")

type Iterator interface {
	Next() (entity.Row, error)
}
