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

// DataBaseAccessObject This struct will be the only db access to the outside world.

// type DBAccess interface {
// 	// track
// 	CreateArtist(ctx context.Context, name string) (*Artist, error)
// 	CreateTrack(
// 		ctx context.Context,
// 		title string,
// 		artistId int,
// 		file string) (track *Track, err error)
// 	BulkCreateTracks(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (rows int64, err error)
// 	SearchTrackByTittle(ctx context.Context, title string) (tracks []*Track, err error)
// 	GetTracks(ctx context.Context, limit int32, offset int32) (tracks []*Track, err error)

// 	// artist
// 	GetArtistById(ctx context.Context, id int) (artist *Artist, err error)

// 	// user
// 	CreateUser(ctx context.Context, username, passwordHash string) (err error)
// 	GetUserWithId(ctx context.Context, id string) (user *User, err error)
// 	GetUserWithUsername(ctx context.Context, username string) (user *User, err error)
// }

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
