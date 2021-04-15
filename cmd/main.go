package main

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/hiepd/galedb/pkg/entity"

	"github.com/hiepd/galedb/pkg/storage"

	"github.com/hiepd/galedb/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	host := ""
	port := 2000
	s, err := server.New(host, port, mockDb())
	logrus.SetLevel(logrus.DebugLevel)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create database server")
	}
	go func() {
		logrus.Infof("Listening on %s:%d", host, port)
		if err := s.Start(); err != nil {
			logrus.WithError(err).Fatal("failed to start database server")
		}
	}()
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
	logrus.Warn("Closing database")
	if err := s.Close(); err != nil {
		logrus.WithError(err).Fatal("failed to close database server")
	}
}

func mockDb() *storage.Database {
	cols := []entity.Column{
		{Kind: reflect.Int, Name: "id"},
		{Kind: reflect.String, Name: "user_type"},
		{Kind: reflect.String, Name: "email"},
		{Kind: reflect.Int, Name: "age"},
	}
	tbl := storage.NewPersisentTable(cols)
	tbl.AddRow(entity.Row{Values: []entity.Value{1, "customer", "customer1@example.com", 24}})
	tbl.AddRow(entity.Row{Values: []entity.Value{2, "driver", "driver2@example.com", 30}})
	tbl.AddRow(entity.Row{Values: []entity.Value{3, "customer", "customer3@example.com", 25}})
	tbl.AddRow(entity.Row{Values: []entity.Value{4, "driver", "driver4@example.com", 31}})
	tbl.AddRow(entity.Row{Values: []entity.Value{5, "customer", "customer5@example.com", 40}})
	return &storage.Database{
		Catalog: map[string]*storage.PersistentTable{
			"users": tbl.(*storage.PersistentTable),
		},
	}
}
