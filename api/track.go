package api

import (
	"auxstream/db"
	fs "auxstream/file_system"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"strconv"
)

// FetchTracksHandler fetch tracks by artist (limit results < 100)
func FetchTracksHandler(c *gin.Context) {
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

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context) {
	form, _ := c.MultipartForm()
	trackTittle := form.Value["title"][0]
	trackArtist := form.Value["artist"][0]
	file := form.File["audio"][0]
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
	fileName, err := fs.Store.Save(raw_bytes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("(store) audio upload failed: %s", err.Error())))
		return
	}
	track, err := db.DAO.CreateTrack(c, trackTittle, trackArtist, fileName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("(db) audio upload failed: %s", err.Error())))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}

// BulkTrackUploadHandler enables bulk track uploads
func BulkTrackUploadHandler(c *gin.Context) {
	form, _ := c.MultipartForm()
	titles := form.Value["track_title"]
	files := form.File["track_files"]
	artistId, err := strconv.Atoi(form.Value["artist_id"][0])

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(fmt.Sprintf("invalid artist_id value: %s", err.Error())))
	}

	fileNames, err := processFiles(files)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("audio upload failed: %s", err.Error())))
		return
	}

	var trackTitles []string
	var filteredFileNames []string
	// filter tracks that failed to upload
	for idx, fileName := range fileNames {
		if fileName != "" {
			trackTitles = append(trackTitles, titles[idx])
			filteredFileNames = append(filteredFileNames, fileName)
		}
	}
	rows, err := db.DAO.BulkCreateTracks(c, trackTitles, artistId, filteredFileNames)

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
