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

	v1 := r.Group("/api/v1")

	// Set up the cookie store for session management
	v1.Use(sessions.Sessions("usersession", cookie.NewStore(sessionSecret)))

	// POST routes
	v1.POST("/upload_track", CookieAuthMiddleware, AddTrackHandler)
	v1.POST("/upload_batch_track", CookieAuthMiddleware, BulkTrackUploadHandler)
	v1.POST("/login", Login)
	v1.POST("/signup", Signup)

	// GET routes
	v1.GET("/search", FetchTracksHandler)
	v1.GET("/logout", Logout)

	return r
}

func SetupTestRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/upload_track", AddTrackHandler)
	r.POST("/upload_batch_track", BulkTrackUploadHandler)
	r.GET("/search", FetchTracksHandler)

	return r
}
