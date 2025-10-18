package main

import (
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"auxstream/internal/http"
	fs "auxstream/internal/storage"
	"auxstream/config"
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load env config: ", err.Error())
	}

	// Initialize database with GORM
	database := db.InitDB(config, context.Background())

	gin.SetMode(config.GinMode)

	rc := cache.NewRedis(&redis.Options{
		Addr: config.RedisAddr,
	})

	err = fs.SetFileStore(config)

	if err != nil {
		log.Fatalf("failed set file store: %s", err.Error())
	}

	server := http.NewServer(http.ServerConfig{
		Cache: rc,
		DB:    database,
		Conf:  config,
	})

	err = server.Run()
	if err != nil {
		log.Fatalf("failed to start server: %s", err.Error())
	}

}
