package storage

import (
	"fmt"
)

type Database struct {
	Name    string
	Catalog map[string]*PersistentTable
}

func (db *Database) GetTable(tableName string) (*PersistentTable, error) {
	t, ok := db.Catalog[tableName]
	if !ok {
		return nil, fmt.Errorf("cannot find table %s in database %s", tableName, db.Name)
	}
	return t, nil
}
