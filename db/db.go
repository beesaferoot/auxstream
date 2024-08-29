package db

import (
	"auxstream/utils"
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DbConn  abstraction over pgx.Conn which allows for mock testing
type DbConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
	Close(ctx context.Context) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

func InitDB(config utils.Config, ctx context.Context) DbConn {
	conn, err := pgx.Connect(ctx, config.DBUrl)

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn
}

func closeDB(conn DbConn, ctx context.Context) {
	err := conn.Close(ctx)
	if err != nil {
		log.Println(err)
	}
}
