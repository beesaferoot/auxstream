package api

import (
	"auxstream/cache"
	"auxstream/db"
	fs "auxstream/file_system"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"time"
)

// FetchTracksByArtistHandler fetch tracks by artist (limit results < 100)
func FetchTracksByArtistHandler(c *gin.Context) {
	artist := c.Query("artist")
	tracks, err := db.DAO.SearchTrackByArtist(c, artist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type FetchTrackQueryParams struct {
	PageSize int8 `form:"pagesize" binding:"gte=0"`
	PageNum  int8 `form:"pagenumber" binding:"gte=1"`
}

// FetchTracksHandler fetch paginated tracks with limit on page size
func FetchTracksHandler(c *gin.Context) {
	var reqParams FetchTrackQueryParams

	fmt.Printf("pagesize: %s\npagenum: %s\n ", c.Query("pagesize"), c.Query("pagenumber"))
	if err := c.ShouldBindQuery(&reqParams); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	limit := reqParams.PageSize
	offset := (reqParams.PageNum - 1) * reqParams.PageSize

	tracks, err := db.DAO.GetTracks(c, limit, offset)
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
	ArtistId int                   `form:"artist_id" binding:"required"`
	Audio    *multipart.FileHeader `form:"audio" binding:"required"`
}

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context) {
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
	raw_file, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("unable to access track audio"))
		return
	}
	raw_bytes := make([]byte, file.Size)
	_, err = raw_file.Read(raw_bytes)
	fileName, err := fs.LStore.Save(raw_bytes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("(store) audio upload failed: %s", err.Error())))
		return
	}

	artist := &db.Artist{Id: trackArtistId}

	ctx := c.Request.Context()
	cacheClient, ok := ctx.Value("cacheClient").(cache.Cache)

	artistCacheKey := "artist-id-" + fmt.Sprintf("%d", trackArtistId)
	// cache client exists
	if ok {
		err = cacheClient.Get(artistCacheKey, artist)
		if err != nil {
			fmt.Printf("(Get artist id from cache) failed: %s\n", err.Error())
			err = nil
		}
	}

	// artist should point to a value from cache if cache hit was successful
	if artist.Name == "" {
		artist, err = db.DAO.GetArtistById(c, trackArtistId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errorResponse(fmt.Sprintf("artist with id (%d) does not exists: %s", trackArtistId, err.Error())))
			return
		}
		_ = cacheClient.Set(artistCacheKey, artist, 10*time.Hour)
	}

	track, err := db.DAO.CreateTrack(c, trackTittle, trackArtistId, fileName)
	if err != nil {
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
	ArtistId int                     `form:"artist_id" binding:"required"`
}

// BulkTrackUploadHandler enables bulk track uploads
func BulkTrackUploadHandler(c *gin.Context) {
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
	rows, err := db.DAO.BulkCreateTracks(c, trackTitles, reqForm.ArtistId, filteredFileNames)

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
	fs.LStore.BulkSave(buf_channel, groupfiles)

	for fileName := range buf_channel {
		fileNames = append(fileNames, fileName)
	}

	return
}
