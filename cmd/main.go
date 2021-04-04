package main

import (
	"os"
	"os/signal"
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
	tbl := storage.NewPersisentTable()
	tbl.AddRow(entity.Row{
		Values: []entity.Value{"hello", int32(1), "you"},
	})
	return &storage.Database{
		Catalog: map[string]*storage.PersistentTable{
			"t1": tbl.(*storage.PersistentTable),
		},
	}
}
