package main

import (
	"auxstream/api"
	"auxstream/db"
	"auxstream/utils"
	"context"
	"log"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("could not load env config: ", err.Error())
	}
	db.DAO = db.New(config, context.Background())
	router := api.SetupRouter(config)
	router.ForwardedByClientIP = true
	err = router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = router.Run(":5009")
	if err != nil {
		log.Fatalln("failed to start server")
	}

}
