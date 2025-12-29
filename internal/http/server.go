package http

import (
	"auxstream/config"
	"auxstream/internal/auth"
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"auxstream/internal/external"
	"auxstream/internal/http/handlers"
	"auxstream/internal/http/middleware"
	"auxstream/internal/logger"
	"auxstream/internal/search"
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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
	db            *gorm.DB
	cache         cache.Cache
	conf          config.Config
	jwtService    *auth.JWTService
	authService   *handlers.AuthService
	searchService *search.Service
	rateLimiter   *middleware.RateLimiter
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

	// Initialize YouTube and SoundCloud clients
	youtubeClient := external.NewYouTubeClient(serverConfig.Conf.YouTubeAPIKey)
	soundcloudClient := external.NewSoundCloudClient(serverConfig.Conf.SoundCloudClientID)

	// Initialize search aggregator
	aggregator := external.NewAggregator(youtubeClient, soundcloudClient, db.NewTrackRepo(serverConfig.DB))
	searchService := search.NewService(aggregator, serverConfig.Cache)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(serverConfig.Cache, middleware.RateLimitConfig{
		MaxRequests: 20,
		Window:      time.Minute,
	})

	return &server{
		db:            serverConfig.DB,
		cache:         serverConfig.Cache,
		conf:          serverConfig.Conf,
		jwtService:    jwtService,
		authService:   authService,
		searchService: searchService,
		rateLimiter:   rateLimiter,
	}
}

func NewMockServer(db *gorm.DB, cache cache.Cache) Server {
	return &server{
		db:    db,
		cache: cache,
	}
}

func (s *server) Run() error {
	environment := "production"
	if s.conf.GinMode == "debug" || s.conf.GinMode == "development" {
		environment = "development"
	}

	if err := logger.InitLogger(environment); err != nil {
		return err
	}
	defer logger.Sync()

	logger.Info("Starting AuxStream server",
		zap.String("environment", environment),
		zap.String("address", s.conf.Addr+":"+s.conf.Port),
		zap.String("gin_mode", s.conf.GinMode),
	)

	router := s.SetupRouter(false)

	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err
	}

	logger.Info("Server listening",
		zap.String("port", s.conf.Port),
	)

	return router.Run(s.conf.Addr + ":" + s.conf.Port)
}

func (s *server) SetupRouter(mock bool) *gin.Engine {
	if mock {
		return s.setupMockRouter()
	}
	return s.setupRouter()
}

func (s *server) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.MaxMultipartMemory = 5 << 20 // 5 miB

	r.Use(middleware.LoggingMiddleware())

	corsConfig := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
	r.Use(corsConfig)
	r.Use(injectCache(s.cache))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", s.healthCheck)

	v1 := r.Group("/api/v1")

	// Authentication routes
	v1.POST("/register", s.authService.Register)
	v1.POST("/login", s.authService.Login)
	v1.POST("/refresh", s.authService.RefreshToken)
	v1.POST("/logout", s.authService.Logout)
	v1.GET("/auth/google", s.authService.GoogleAuth)
	v1.GET("/auth/google/callback", s.authService.GoogleCallback)

	// Artist Management Routes
	artists := v1.Group("/artists")
	{
		// Public routes
		artists.GET("/:id", func(c *gin.Context) {
			handlers.GetArtistByIdHandler(c, db.NewArtistRepo(s.db))
		})
		artists.GET("/:id/tracks", func(c *gin.Context) {
			handlers.GetArtistTracksHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
		})
		artists.GET("/search", func(c *gin.Context) {
			handlers.FetchTracksByArtistHandler(c, db.NewTrackRepo(s.db))
		})

		// Protected routes
		artists.POST("", s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
			handlers.CreateArtistHandler(c, db.NewArtistRepo(s.db))
		})
	}

	// Track Management Routes
	tracks := v1.Group("/tracks")
	{
		// Public routes
		tracks.GET("", func(c *gin.Context) {
			handlers.FetchTracksHandler(c, db.NewTrackRepo(s.db))
		})
		tracks.GET("/:id", func(c *gin.Context) {
			handlers.GetTrackByIDHandler(c, db.NewTrackRepo(s.db))
		})

		// Track playback recording (no auth required to allow anonymous plays)
		tracks.POST("/play", func(c *gin.Context) {
			handlers.TrackPlayHandler(c, db.NewTrackRepo(s.db))
		})

		// Protected routes with rate limiting
		tracks.POST("", s.rateLimiter.Middleware(), s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
			handlers.AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
		})
		tracks.POST("/bulk", s.rateLimiter.Middleware(), s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
			handlers.BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
		})
	}

	// Legacy upload endpoints (for backward compatibility)
	v1.POST("/upload_track", s.rateLimiter.Middleware(), s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.AddTrackHandler(c, db.NewTrackRepo(s.db), db.NewArtistRepo(s.db))
	})
	v1.POST("/upload_batch_track", s.rateLimiter.Middleware(), s.jwtService.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.BulkTrackUploadHandler(c, db.NewTrackRepo(s.db))
	})

	// Search Routes
	v1.GET("/search", s.rateLimiter.Middleware(), func(c *gin.Context) {
		handlers.SearchHandler(c, s.searchService)
	})

	// Static file serving
	v1.Static("/serve", "./uploads")

	return r
}

func (s *server) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"service":   "auxstream",
		"timestamp": time.Now().Unix(),
	})
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

func createHTTPHandler(funcHandler func(*gin.Context, ...any), repos ...any) gin.HandlerFunc {
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
