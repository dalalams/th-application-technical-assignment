package database

import (
	"context"
	"th-application-technical-assignment/sqlc"
)

type Connection interface {
	DBTX() sqlc.DBTX
	Close(ctx context.Context)
}

type Store struct {
	conn    Connection
	Queries sqlc.Querier
}

func New(ctx context.Context, conn Connection) *Store {
	return &Store{conn: conn, Queries: sqlc.New(conn.DBTX())}
}

func (s *Store) Close(ctx context.Context) {
	s.conn.Close(ctx)
}
