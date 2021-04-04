package server

import (
	"github.com/hiepd/galedb/pkg/server/pgwire"
	"github.com/hiepd/galedb/pkg/sql/parser"
	"github.com/hiepd/galedb/pkg/storage"
)

type Server struct {
	PgWire *pgwire.Server
	Parser *parser.Parser
}

func New(host string, port int, defaultDb *storage.Database) (*Server, error) {
	ps, err := pgwire.NewServer(port, host, parser.New(), defaultDb)
	if err != nil {
		return nil, err
	}
	return &Server{
		PgWire: ps,
	}, nil
}

func (s *Server) Start() error {
	return s.PgWire.Start()
}

func (s *Server) Close() error {
	return s.PgWire.Close()
}
