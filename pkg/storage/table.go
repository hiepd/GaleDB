package storage

type Table interface {
	IsPersistent() bool
}
