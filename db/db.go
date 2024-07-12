package db

import (
	"auxstream/utils"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gopkg.in/validator.v2"
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

type DBAccess interface {
	// track
	CreateArtist(ctx context.Context, name string) (*Artist, error)
	CreateTrack(
		ctx context.Context,
		title string,
		artistId int,
		file string) (track *Track, err error)
	BulkCreateTracks(ctx context.Context, trackTitles []string, artistId int, fileNames []string) (rows int64, err error)
	SearchTrackByTittle(ctx context.Context, title string) (tracks []*Track, err error)
	GetTracks(ctx context.Context, limit int32, offset int32) (tracks []*Track, err error)

	// user
	CreateUser(ctx context.Context, username, passwordHash string) (err error)
	GetUserWithId(ctx context.Context, id string) (user *User, err error)
	GetUserWithUsername(ctx context.Context, username string) (user *User, err error)
}

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

func New(config utils.Config, context context.Context) *DataBaseAccessObject {
	conn, err := pgx.Connect(context, config.DBUrl)

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

func (dao *DataBaseAccessObject) CreateArtist(ctx context.Context, name string) (artist *Artist, err error) {
	artist = &Artist{Name: name}
	if err := validator.Validate(artist); err != nil {
		return nil, err
	}
	err = artist.Commit(ctx)
	return artist, err
}

func (dao *DataBaseAccessObject) CreateTrack(
	ctx context.Context,
	title string,
	artistId int,
	file string) (track *Track, err error) {
	track = &Track{Title: title, File: file, ArtistId: artistId}
	if err := validator.Validate(track); err != nil {
		return nil, err
	}
	err = track.Commit(ctx)
	return track, err
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

func (dao *DataBaseAccessObject) GetTracks(ctx context.Context, limit int32, offset int32) (tracks []*Track, err error) {
	return GetTracks(ctx, limit, offset)
}

func (dao *DataBaseAccessObject) CreateUser(ctx context.Context, username, passwordHash string) (err error) {
	user := &User{Username: username, PasswordHash: passwordHash}
	if err := validator.Validate(user); err != nil {
		return err
	}
	return user.Commit(ctx)
}

func (dao *DataBaseAccessObject) GetUserWithId(ctx context.Context, id string) (user *User, err error) {
	return GetUserById(ctx, id)
}

func (dao *DataBaseAccessObject) GetUserWithUsername(ctx context.Context, username string) (user *User, err error) {
	return GetUserByUsername(ctx, username)
}
