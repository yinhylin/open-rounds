package main

import (
	"log"
	"os"
	"rounds/server"
)

func main() {
	if err := server.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
