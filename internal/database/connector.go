package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type ConnPool interface {
	GetSchema() string
	Close()
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close()
}

type connPool struct {
	db     PgxIface
	schema string
}

func NewConnPool(db PgxIface, schema string) ConnPool {
	return &connPool{
		db:     db,
		schema: schema,
	}
}

func (p *connPool) GetSchema() string {
	return p.schema
}

func (p *connPool) Close() {
	p.db.Close()
}

func (p *connPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.db.Begin(ctx)
}
