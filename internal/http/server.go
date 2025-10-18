package http

import (
	"auxstream/internal/auth"
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"auxstream/internal/http/handlers"
	"auxstream/config"
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server interface {
	Run() error
	SetupRouter(mock bool) *gin.Engine
}

type ServerConfig struct {
	DB    *gorm.DB
	Cache cache.Cache
	Conf  config.Config
}

type server struct {
	db          *gorm.DB
	cache       cache.Cache
	conf        config.Config
	jwtService  *auth.JWTService
	authService *handlers.AuthService
}

func NewServer(serverConfig ServerConfig) Server {
	// Initialize JWT service
	jwtService := auth.NewJWTService(
		serverConfig.Conf.JWTSecret,
		time.Hour,      // Access token TTL
		24*time.Hour*7, // Refresh token TTL (7 days)
	)

	// Initialize refresh token service
	refreshService := auth.NewRefreshTokenService(serverConfig.Cache, jwtService)

	// Initialize OAuth service
	oauthService := auth.NewOAuthService(
		serverConfig.Conf.GoogleClientID,
		serverConfig.Conf.GoogleClientSecret,
		serverConfig.Conf.GoogleRedirectURL,
		db.NewUserRepo(serverConfig.DB),
	)

	// Initialize auth service
	authService := handlers.NewAuthService(
		db.NewUserRepo(serverConfig.DB),
		jwtService,
		refreshService,
		oauthService,
	)

	return &server{
		db:          serverConfig.DB,
		cache:       serverConfig.Cache,
		conf:        serverConfig.Conf,
		jwtService:  jwtService,
		authService: authService,
	}
}

func NewMockServer(db *gorm.DB, cache cache.Cache) Server {
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

	// Allow cors origin
	config := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
	r.Use(config)
	r.Use(injectCache(s.cache))
	v1 := r.Group("/api/v1")

	// Note: Session management removed - using JWT tokens instead

	// Authentication routes
	v1.POST("/register", s.authService.Register)
	v1.POST("/login", s.authService.Login)
	v1.POST("/refresh", s.authService.RefreshToken)
	v1.POST("/logout", s.authService.Logout)
	v1.GET("/auth/google", s.authService.GoogleAuth)
	v1.GET("/auth/google/callback", s.authService.GoogleCallback)

	// Protected routes
	v1.POST("/upload_track", s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
	})
	v1.POST("/upload_batch_track", s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
	})

	// GET routes
	v1.GET("/tracks", func(c *gin.Context) {
		handlers.FetchTracksHandler(c, db.NewTrackRepo(s.db))
	})
	v1.GET("/search", func(c *gin.Context) {
		handlers.FetchTracksByArtistHandler(c, db.NewTrackRepo(s.db))
	})
	v1.Static("/serve", "./uploads")

	return r
}

func (s *server) setupMockRouter() *gin.Engine {
	r := gin.Default()
	r.Use(injectCache(s.cache))
	r.POST("/upload_track", func(c *gin.Context) {
		handlers.AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
	})
	r.POST("/upload_batch_track", func(c *gin.Context) {
		handlers.BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
	})
	r.GET("/tracks", func(c *gin.Context) {
		handlers.FetchTracksHandler(c, db.NewTrackRepo(s.db))
	})
	r.GET("/search", func(c *gin.Context) {
		handlers.FetchTracksByArtistHandler(c, db.NewTrackRepo(s.db))
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
