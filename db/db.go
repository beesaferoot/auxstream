package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

// dbConn  abstraction over pgx.Conn which allows for mock testing
type dbConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
	Close(ctx context.Context) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// DataBaseAccessObject This struct will be the only db access to the outside world.
type DataBaseAccessObject struct {
	conn dbConn
}

var DAO *DataBaseAccessObject

func (dao *DataBaseAccessObject) setupDB() {

}

func (dao *DataBaseAccessObject) close(context context.Context) {
	err := dao.conn.Close(context)
	if err != nil {
		log.Println(err)
	}
}

type DBconfig struct {
	Url string
}

func New(config DBconfig, context context.Context) *DataBaseAccessObject {
	conn, err := pgx.Connect(context, config.Url)

	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			return nil
		}
		os.Exit(1)
	}
	return &DataBaseAccessObject{conn: conn}
}

func NewWithMockConn(conn dbConn) *DataBaseAccessObject {
	return &DataBaseAccessObject{conn: conn}
}

/*
	DB operations
*/

func (dao *DataBaseAccessObject) CreateTrack(
	ctx context.Context,
	title string,
	artistName string,
	file string) (track *Track, err error) {
	artist := &Artist{Name: artistName}
	trx, err := dao.conn.Begin(ctx)
	// Rollback db transaction on failure
	defer func() {
		if err != nil {
			_ = trx.Rollback(ctx)
		}
	}()
	if err != nil {
		return nil, err
	}
	err = artist.Commit(ctx, trx)
	if err != nil {
		return nil, err
	}
	track = &Track{Title: title, File: file, ArtistId: artist.Id}
	err = track.Commit(ctx, trx)
	return track, nil
}

func (dao *DataBaseAccessObject) BulkCreateTracks(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (rows int64, err error) {
	return BulkCreateTrack(ctx, trackTitles, artistId, fileNames)
}

func (dao *DataBaseAccessObject) SearchTrackByTittle(ctx context.Context, title string) (tracks []*Track, err error) {
	return GetTrackByTitle(ctx, title)
}

func (dao *DataBaseAccessObject) SearchTrackByArtist(ctx context.Context,
	artist string) (tracks []*Track, err error) {
	return GetTrackByArtist(ctx, artist)
}
