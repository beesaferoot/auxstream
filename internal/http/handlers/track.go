package handlers

import (
	"auxstream/internal/cache"
	"auxstream/internal/db"
	fs "auxstream/internal/storage"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// FetchTracksByArtistHandler fetch tracks by artist (limit results < 100)
func FetchTracksByArtistHandler(c *gin.Context, r db.TrackRepo) {
	artist := c.Query("artist")
	tracks, err := r.GetTrackByArtist(c, artist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type FetchTrackQueryParams struct {
	PageSize int `form:"pagesize" binding:"gte=0"`
	PageNum  int `form:"pagenumber" binding:"gte=1"`
}

// FetchTracksHandler fetch paginated tracks with limit on page size
func FetchTracksHandler(c *gin.Context, r db.TrackRepo) {
	var reqParams FetchTrackQueryParams

	if err := c.ShouldBindQuery(&reqParams); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	limit := reqParams.PageSize
	offset := (reqParams.PageNum - 1) * reqParams.PageSize

	tracks, err := r.GetTracks(c, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type AddTrackForm struct {
	Title    string                `form:"title" binding:"required"`
	ArtistId uuid.UUID             `form:"artist_id" binding:"required"`
	Audio    *multipart.FileHeader `form:"audio" binding:"required"`
}

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context, r db.TrackRepo, artistRepo db.ArtistRepo) {
	var reqForm AddTrackForm
	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	trackTittle := reqForm.Title
	trackArtistId := reqForm.ArtistId

	file := reqForm.Audio

	if file.Size <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("audio for track not found"))
		return
	}

	if !strings.HasSuffix(file.Filename, ".mp3") {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("invalid audio format use mp3 instead"))
		return
	}

	raw_file, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("unable to access track audio"))
		return
	}
	raw_bytes := make([]byte, file.Size)
	_, err = raw_file.Read(raw_bytes)
	filePath, err := fs.Store.Save(raw_bytes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("(store) audio upload failed: %s", err.Error())))
		return
	}

	artist := &db.Artist{ID: trackArtistId}

	ctx := c.Request.Context()
	cacheClient, ok := ctx.Value("cacheClient").(cache.Cache)

	artistCacheKey := "artist-id-" + fmt.Sprintf("%d", trackArtistId)
	// cache client exists
	if ok {
		err = cacheClient.Get(artistCacheKey, &cache.Cacheable[db.Artist]{Value: artist})
		if err != nil {
			log.Printf("(Get artist id from cache) failed: %s\n", err.Error())
			err = nil
		}
	}

	// artist should point to a value from cache if cache hit was successful
	if artist.Name == "" {
		artist, err = artistRepo.GetArtistById(c, trackArtistId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errorResponse(fmt.Sprintf("artist with id (%d) does not exists: %s", trackArtistId, err.Error())))
			return
		}
		_ = cacheClient.Set(artistCacheKey, &cache.Cacheable[db.Artist]{Value: artist}, 10*time.Hour)
	}

	track, err := r.CreateTrack(c, trackTittle, trackArtistId, filePath)
	if err != nil {
		fmt.Printf("Artist: %v\n", artist)
		fmt.Printf("Error: %s\n", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("(db) audio upload failed: %s", err.Error())))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}

type BulkTrackUploadForm struct {
	Titles   []string                `form:"track_titles" binding:"required"`
	Files    []*multipart.FileHeader `form:"track_files" binding:"required"`
	ArtistId uuid.UUID                     `form:"artist_id" binding:"required"`
}

// BulkTrackUploadHandler enables bulk track uploads
func BulkTrackUploadHandler(c *gin.Context, r db.TrackRepo) {
	var reqForm BulkTrackUploadForm

	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	fileNames, err := processFiles(reqForm.Files)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("audio upload failed: %s", err.Error())))
		return
	}

	var trackTitles []string
	var filteredFileNames []string
	// filter tracks that failed to upload
	for idx, fileName := range fileNames {
		if fileName != "" {
			trackTitles = append(trackTitles, reqForm.Titles[idx])
			filteredFileNames = append(filteredFileNames, fileName)
		}
	}
	rows, err := r.BulkCreateTracks(c, trackTitles, reqForm.ArtistId, filteredFileNames)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("audio upload failed: %s", err.Error())))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": map[string]any{
			"saved": filteredFileNames,
			"rows":  rows,
		},
	})
}

func processFiles(files []*multipart.FileHeader) (fileNames []string, err error) {
	var groupfiles [][]byte

	for _, file := range files {
		raw_file, err := file.Open()
		if err != nil {
			err = nil
			groupfiles = append(groupfiles, []byte{})
			continue
		}
		raw_bytes := make([]byte, file.Size)
		_, err = raw_file.Read(raw_bytes)
		if err != nil {
			err = nil
			groupfiles = append(groupfiles, []byte{})
			continue
		}
		groupfiles = append(groupfiles, raw_bytes)
	}

	buf_channel := make(chan string, len(groupfiles))

	// concurrently write files to disk
	fs.Store.BulkSave(buf_channel, groupfiles)

	for fileName := range buf_channel {
		fileNames = append(fileNames, fileName)
	}

	return
}
