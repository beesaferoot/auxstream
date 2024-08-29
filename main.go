package main

import (
	"auxstream/api"
	"auxstream/cache"
	"auxstream/db"
	fs "auxstream/file_system"
	"auxstream/utils"
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load env config: ", err.Error())
	}
	dB := db.InitDB(config, context.Background())
	gin.SetMode(config.GinMode)

	rc := cache.NewRedis(&redis.Options{
		Addr: config.RedisAddr,
	})

	err = fs.SetFileStore(config)

	if err != nil {
		log.Fatalf("failed set file store: %s", err.Error())
	}

	server := api.NewServer(api.ServerConfig{
		Cache: rc,
		DB:    dB,
		Conf:  config,
	})

	err = server.Run()
	if err != nil {
		log.Fatalf("failed to start server: ", err.Error())
	}

}
