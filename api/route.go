package api

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// POST routes
	r.POST("/upload_track", AddTrackHandler)
	// GET routes
	r.GET("/search", FetchTracksHandler)
	return r
}
