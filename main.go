package main

import (
	"auxstream/api"
)

func main() {
	router := api.SetupRouter()
	router.Run(":5009")
}
