package api

import (
	"github.com/gin-gonic/gin"
)

func SetupTestRouter() *gin.Engine {
	r := gin.Default()
	// r.POST("/upload_track", AddTrackHandler)
	// r.POST("/upload_batch_track", BulkTrackUploadHandler)
	// r.GET("/tracks", FetchTracksHandler)
	// r.GET("/search", FetchTracksByArtistHandler)

	return r
}
