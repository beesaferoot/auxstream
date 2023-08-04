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
	router := api.SetupRouter()
	err = router.Run(":5009")
	if err != nil {
		log.Fatalln("failed to start server")
	}

}
