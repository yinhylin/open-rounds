package main

import (
	"log"
	"os"

	"github.com/sailormoon/open-rounds/server"
)

func main() {
	if err := server.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
