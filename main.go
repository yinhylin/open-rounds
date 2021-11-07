package main

import (
	"context"
	"log"
	"rounds/client"
	"rounds/server"
	"rounds/utils"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"nhooyr.io/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	cfg, err := utils.ReadTOML("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	resolutionConfig := cfg.UI.Resolution
	log.Printf("%+v", resolutionConfig)

	ebiten.SetWindowSize(resolutionConfig.X, resolutionConfig.Y)
	ebiten.SetWindowTitle("Open ROUNDS")

	player := client.NewLocalPlayer()
	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "ws://localhost:4242", nil)
	if err != nil {
		log.Printf("Encountered err: %v. Trying to spin up server manually\n", err)
		// Try to spin up the server if we fail to connect.
		go server.Run()

		// TODO: Should have a good way of testing if the server is up.
		time.Sleep(50 * time.Millisecond)
		c, _, err = websocket.Dial(ctx, "ws://localhost:4242", nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer c.Close(websocket.StatusInternalError, "")

	game := client.NewGame(player)

	go game.ReadMessages(ctx, c)
	go player.WriteMessages(ctx, c)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
