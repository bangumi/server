package main

import (
	"log"

	"github.com/bangumi/server/internal/search/cron"
)

func main() {
	if err := cron.Start(); err != nil {
		log.Fatalln(err)
	}
}
