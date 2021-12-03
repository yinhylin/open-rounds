package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/sailormoon/open-rounds/client"
	"github.com/sailormoon/open-rounds/server"

	"github.com/hajimehoshi/ebiten/v2"
	"nhooyr.io/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	if len(os.Args) > 1 && os.Args[1] == "server" {
		if err := server.NewServer().Run(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	host := "localhost:4242"
	if len(os.Args) > 1 {
		host = os.Args[1]
	}

	assets, err := client.LoadAssets()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Open ROUNDS")

	game := client.NewGame(assets)

	ctx := context.Background()
	errc := make(chan error)
	go func() {
		c, _, err := websocket.Dial(ctx, "ws://"+host, nil)
		if err != nil {
			log.Printf("Encountered err: %v. Trying to spin up server manually\n", err)

			server := server.NewServer()
			// Try to spin up the server if we fail to connect.
			go func() {
				if err := server.Run([]string{}); err != nil {
					log.Fatal(err)
				}
				log.Fatal("server shutdown")
			}()

			server.WaitForRunning()
			c, _, err = websocket.Dial(ctx, "ws://localhost:4242", nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		go game.ReadMessages(ctx, c)
		go game.WriteMessages(ctx, c)
		err = <-errc
		c.Close(websocket.StatusInternalError, err.Error())
	}()

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	if err := ebiten.RunGame(game); err != nil {
		errc <- err
		log.Fatal(err)
	}
}
