package db

import (
	"context"
)

type Database interface {
	Queries() interface{}
	Close()
	Ping(ctx context.Context) error
}

type DatabaseWithTransaction interface {
	Database
	Begin(ctx context.Context) (interface{}, error)
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Queries() interface{}
}
