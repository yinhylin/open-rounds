package main

import (
	"log"
	"os"

	"github.com/sailormoon/open-rounds/server"
)

func main() {
	if err := server.NewServer().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
