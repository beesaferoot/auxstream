package main

import (
	"auxstream/config"
	"auxstream/internal/cache"
	"auxstream/internal/db"
	"auxstream/internal/http"
	fs "auxstream/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load env config: ", err.Error())
	}

	database := db.InitDB(conf)

	gin.SetMode(conf.GinMode)

	rc := cache.NewRedis(&redis.Options{
		Addr: conf.RedisAddr,
	})

	if err = fs.SetFileStore(conf); err != nil {
		log.Fatalf("failed to set file store: %s", err.Error())
	}

	server := http.NewServer(http.ServerConfig{
		Cache: rc,
		DB:    database,
		Conf:  conf,
	})

	if err = server.Run(); err != nil {
		log.Fatalf("failed to start server: %s", err.Error())
	}
}
