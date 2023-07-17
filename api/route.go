package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"os"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	sessionSecret := []byte(os.Getenv("sessionsecret"))

	// Set up the cookie store for session management
	r.Use(sessions.Sessions("usersession", cookie.NewStore(sessionSecret)))

	// POST routes
	r.POST("/upload_track", CookieAuthMiddleware, AddTrackHandler)
	r.POST("/upload_batch_track", CookieAuthMiddleware, BulkTrackUploadHandler)
	r.POST("/login", Login)
	r.POST("/signup", Signup)

	// GET routes
	r.GET("/search", FetchTracksHandler)
	r.GET("/logout", Logout)

	return r
}

func SetupTestRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/upload_track", AddTrackHandler)
	r.POST("/upload_batch_track", BulkTrackUploadHandler)
	r.GET("/search", FetchTracksHandler)

	return r
}
