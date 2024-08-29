package api

import (
	"auxstream/cache"
	"auxstream/db"
	"auxstream/utils"
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Server interface {
	Run() error
	SetupRouter(mock bool) *gin.Engine
}

type ServerConfig struct {
	DB    db.DbConn
	Cache cache.Cache
	Conf  utils.Config
}

type server struct {
	db    db.DbConn
	cache cache.Cache
	conf  utils.Config
}

func NewServer(serverConfig ServerConfig) Server {
	return &server{
		db:    serverConfig.DB,
		cache: serverConfig.Cache,
		conf:  serverConfig.Conf,
	}
}

func NewMockServer(db db.DbConn, cache cache.Cache) Server {
	return &server{
		db:    db,
		cache: cache,
	}
}

func (s *server) Run() error {
	router := s.SetupRouter(false)

	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err
	}

	return router.Run(s.conf.Addr + ":" + s.conf.Port)
}

func (s *server) SetupRouter(mock bool) *gin.Engine {
	if mock {
		return s.setupMockRouter()
	}
	return s.setupRouter()
}

func (s *server) setupRouter() *gin.Engine {
	r := gin.Default()
	
	r.MaxMultipartMemory = 5 << 20 // 5 miB
	
	sessionSecret := []byte(s.conf.SessionString)
	// Allow cors origin
	config := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
	r.Use(config)
	r.Use(injectCache(s.cache))
	v1 := r.Group("/api/v1")

	// Set up the cookie store for session management
	v1.Use(sessions.Sessions("usersession", cookie.NewStore(sessionSecret)))

	// POST routes
	v1.POST("/upload_track", CookieAuthMiddleware, func(c *gin.Context) {
		AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
	})
	v1.POST("/upload_batch_track", CookieAuthMiddleware, func(c *gin.Context) {
		BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
	})
	v1.POST("/login", func(c *gin.Context) {
		Login(c, db.NewUserRepo(s.db))
	})
	v1.POST("/signup", func(c *gin.Context) {
		Signup(c, db.NewUserRepo(s.db))
	})

	// GET routes
	v1.GET("/tracks", func(c *gin.Context) {
		FetchTracksHandler(c, db.NewTrackRepo(s.db))
	})
	v1.GET("/search", func(c *gin.Context) {
		FetchTracksByArtistHandler(c, db.NewTrackRepo(s.db))
	})
	v1.GET("/logout", func(c *gin.Context) {
		Logout(c)
	})
	v1.Static("/serve", "./uploads")

	return r
}

func (s *server) setupMockRouter() *gin.Engine {
	r := gin.Default()
	r.Use(injectCache(s.cache))
	r.POST("/upload_track", func(c *gin.Context) {
		AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
	})
	r.POST("/upload_batch_track", func(c *gin.Context) {
		BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
	})
	r.GET("/tracks", func(c *gin.Context) {
		FetchTracksHandler(c, db.NewTrackRepo(s.db))
	})
	r.GET("/search", func(c *gin.Context) {
		FetchTracksByArtistHandler(c, db.NewTrackRepo(s.db))
	})

	return r
}

func createHTTPHandler(funcHandler func(*gin.Context, ...interface{}), repos ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		funcHandler(c, repos...)
	}
}

func injectCache(cache cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "cacheClient", cache)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
