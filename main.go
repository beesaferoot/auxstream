package main

import (
	"auxstream/api"
	"auxstream/cache"
	"auxstream/db"
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

	server := api.NewServer(api.ServerConfig{
		Cache: rc,
		DB:    dB,
		Conf:  config,
	})

	err = server.Run()
	if err != nil {
		log.Fatalln("failed to start server")
	}

}
