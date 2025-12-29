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

// GetTrackByIDHandler fetches a single track by ID
func GetTrackByIDHandler(c *gin.Context, r db.TrackRepo) {
	trackIdStr := c.Param("id")
	trackId, err := uuid.Parse(trackIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track ID format"))
		return
	}

	track, err := r.GetTrackByID(c, trackId)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("track not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": track,
	})
}

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
	PageSize int    `form:"pagesize" binding:"gte=0"`
	PageNum  int    `form:"pagenumber" binding:"gte=1"`
	Sort     string `form:"sort"` // "trending", "recent", or default
	Days     int    `form:"days"` // For trending within last N days (0 = all time)
}

// FetchTracksHandler fetch paginated tracks with limit on page size and sorting options
func FetchTracksHandler(c *gin.Context, r db.TrackRepo) {
	var reqParams FetchTrackQueryParams

	if err := c.ShouldBindQuery(&reqParams); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	limit := reqParams.PageSize
	offset := (reqParams.PageNum - 1) * reqParams.PageSize

	var tracks []*db.Track
	var err error

	// Choose sorting method based on query parameter
	switch reqParams.Sort {
	case "trending":
		days := reqParams.Days
		if days == 0 {
			days = 30 // Default to last 30 days for trending
		}
		tracks, err = r.GetTrendingTracks(c, limit, offset, days)
	case "recent":
		tracks, err = r.GetRecentTracks(c, limit, offset)
	default:
		// Default behavior - just get tracks (no special sorting)
		tracks, err = r.GetTracks(c, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tracks,
	})

}

type AddTrackForm struct {
	Title     string                `form:"title" binding:"required"`
	ArtistId  string                `form:"artist_id" binding:"required"`
	Audio     *multipart.FileHeader `form:"audio" binding:"required"`
	Duration  int                   `form:"duration"`  // Optional: duration in seconds
	Thumbnail string                `form:"thumbnail"` // Optional: thumbnail URL or path
}

// AddTrackHandler add track to the system
func AddTrackHandler(c *gin.Context, r db.TrackRepo, artistRepo db.ArtistRepo) {
	var reqForm AddTrackForm
	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	trackTittle := reqForm.Title

	trackArtistId, err := uuid.Parse(reqForm.ArtistId)

	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Sprintf("artist id should be a valid uuid string not %s", reqForm.ArtistId)))
	}

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
	if _, readErr := raw_file.Read(raw_bytes); readErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse("unable to read track audio"))
		return
	}
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
		err = cacheClient.Get(artistCacheKey, artist)
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
		_ = cacheClient.Set(artistCacheKey, artist, 10*time.Hour)
	}

	track, err := r.CreateTrack(c, trackTittle, trackArtistId, filePath, reqForm.Duration, reqForm.Thumbnail)
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
	ArtistId string                  `form:"artist_id" binding:"required"`
}

// BulkTrackUploadHandler enables bulk track uploads
func BulkTrackUploadHandler(c *gin.Context, r db.TrackRepo) {
	var reqForm BulkTrackUploadForm
	var artistId uuid.UUID

	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	artistId, err := uuid.Parse(reqForm.ArtistId)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Sprintf("invalid artist id: %s", reqForm.ArtistId)))
		return
	}

	fileMetas, err := processFiles(reqForm.Files, reqForm.Titles)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse(fmt.Sprintf("audio upload failed: %s", err.Error())))
		return
	}

	var trackTitles []string
	var filteredFileNames = map[string]string{}
	// filter tracks that failed to upload
	for title, fileMeta := range fileMetas {
		if fileMeta.Name != "" {
			trackTitles = append(trackTitles, title)
			filteredFileNames[title] = fileMeta.Name
		}
	}
	rows, err := r.BulkCreateTracks(c, trackTitles, artistId, filteredFileNames)

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

func processFiles(files []*multipart.FileHeader, titles []string) (fileMeta map[string]fs.FileMeta, err error) {
	var groupFiles []fs.FileMeta
	var groupTitles = map[string]string{}
	for idx, file := range files {
		raw_file, fileErr := file.Open()
		groupTitles[file.Filename] = titles[idx]
		if fileErr != nil {
			groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: []byte{}})
			continue
		}
		raw_bytes := make([]byte, file.Size)
		if _, readErr := raw_file.Read(raw_bytes); readErr != nil {
			groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: []byte{}})
			continue
		}
		groupFiles = append(groupFiles, fs.FileMeta{AudioTitle: file.Filename, Content: raw_bytes[:file.Size]})
	}

	buf_channel := make(chan fs.FileMeta, len(groupFiles))

	// concurrently write files to disk
	fs.Store.BulkSave(buf_channel, groupFiles)

	fileMeta = map[string]fs.FileMeta{}
	for fmeta := range buf_channel {
		if fmeta.Name != "" {
			fileMeta[groupTitles[fmeta.AudioTitle]] = fmeta
		}
	}

	return
}

type TrackPlayRequest struct {
	TrackID        string `json:"track_id" binding:"required"`
	DurationPlayed int    `json:"duration_played"` // Optional: how long the user listened (seconds)
}

// TrackPlayHandler records a track play/listen event
func TrackPlayHandler(c *gin.Context, r db.TrackRepo) {
	var req TrackPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	trackId, err := uuid.Parse(req.TrackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("invalid track ID"))
		return
	}

	// Increment the track's play count
	if err := r.IncrementPlayCount(c, trackId); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("failed to record play"))
		return
	}

	// Optionally record in playback history if user is authenticated
	userIdValue, exists := c.Get("user_id")
	if exists {
		if userId, ok := userIdValue.(uuid.UUID); ok {
			duration := req.DurationPlayed
			if duration == 0 {
				duration = 30 // Default minimum listening time
			}
			_ = r.RecordPlayback(c, userId, trackId, duration)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "play recorded",
	})
}
