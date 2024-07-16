package main

import (
	"auxstream/api"
	"auxstream/cache"
	"auxstream/db"
	"auxstream/utils"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load env config: ", err.Error())
	}
	db.DAO = db.New(config, context.Background())
	gin.SetMode(config.GinMode)

	rc := cache.NewRedis(&redis.Options{
		Addr: config.RedisAddr,
	})
	router := api.SetupRouter(config, rc)

	router.ForwardedByClientIP = true
	err = router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = router.Run(config.Addr + ":" + config.Port)
	if err != nil {
		log.Fatalln("failed to start server")
	}

}
