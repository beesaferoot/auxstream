package main

import (
	"auxstream/api"
	"auxstream/db"
	"context"
	"os"
)

func main() {
	db.DAO = db.New(db.DBconfig{Url: os.Getenv("DATABASE_URL")}, context.Background())
	router := api.SetupRouter()
	router.Run(":5009")

}
