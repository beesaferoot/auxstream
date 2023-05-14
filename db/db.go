package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

/*
	Database Access Object:
		This struct will be the only db access to the outside world.
*/
type DataBaseAccessObject struct {
	conn *pgx.Conn
}

var DAO = New(DBconfig{Url: os.Getenv("DATABASE_URL")}, context.Background())

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
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

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
	// TODO
	return track, nil
}

func (dao *DataBaseAccessObject) BulkCreateTracks(tracks []*Track) (err error) {
	// TODO
	return nil
}

func (dao *DataBaseAccessObject) SearchTrackByTittle(ctx context.Context) (tracks []*Track, err error) {
	// TODO
	return tracks, nil
}

func (dao *DataBaseAccessObject) SearchTrackByArtistName(ctx context.Context) (tracks []*Track, err error) {
	// TODO
	return tracks, nil
}
