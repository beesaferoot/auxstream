package api

import (
	"auxstream/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func SetupRouter(envConfig utils.Config) *gin.Engine {
	r := gin.Default()

	sessionSecret := []byte(envConfig.SessionString)
	// Allow cors origin
	config := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
	r.Use(config)
	v1 := r.Group("/api/v1")

	// Set up the cookie store for session management
	v1.Use(sessions.Sessions("usersession", cookie.NewStore(sessionSecret)))

	// POST routes
	v1.POST("/upload_track", CookieAuthMiddleware, AddTrackHandler)
	v1.POST("/upload_batch_track", CookieAuthMiddleware, BulkTrackUploadHandler)
	v1.POST("/login", Login)
	v1.POST("/signup", Signup)

	// GET routes
	v1.GET("/tracks", FetchTracksHandler)
	v1.GET("/search", FetchTracksByArtistHandler)
	v1.GET("/logout", Logout)
	v1.Static("/serve", "./uploads")

	return r
}

func SetupTestRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/upload_track", AddTrackHandler)
	r.POST("/upload_batch_track", BulkTrackUploadHandler)
	r.GET("/tracks", FetchTracksHandler)
	r.GET("/search", FetchTracksByArtistHandler)

	return r
}
