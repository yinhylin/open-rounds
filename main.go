package main

import (
	"context"
	"log"
	"os"
	"rounds/client"
	"rounds/server"

	"github.com/hajimehoshi/ebiten/v2"
	"nhooyr.io/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	if len(os.Args) > 1 && os.Args[1] == "server" {
		if err := server.Run(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	assets, err := client.LoadAssets()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(1280, 620)
	ebiten.SetWindowTitle("Open ROUNDS")

	game := client.NewGame(assets)

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "ws://44.241.110.166:4242", nil)
	if err != nil {
		log.Fatal(err)
		/*
			log.Printf("Encountered err: %v. Trying to spin up server manually\n", err)

			// Try to spin up the server if we fail to connect.
					go func() {
						if err := server.Run([]string{}); err != nil {
							log.Fatal(err)
						}
						log.Fatal("server shutdown")
					}()

				// TODO: Should have a good way of testing if the server is up.
				time.Sleep(50 * time.Millisecond)
				c, _, err = websocket.Dial(ctx, "ws://localhost:4242", nil)
				if err != nil {
					log.Fatal(err)
				}
		*/
	}
	defer c.Close(websocket.StatusInternalError, "")

	go game.ReadMessages(ctx, c)
	go game.WriteMessages(ctx, c)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
