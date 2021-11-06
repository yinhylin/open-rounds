package main

import (
	"context"
	"log"
	"os"
	"rounds/client"
	"rounds/server"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pelletier/go-toml/v2"
	"nhooyr.io/websocket"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	// Load Configs
	var cfg client.Config
	cfg_file, file_err := os.ReadFile("config.toml")
	if file_err != nil {
		panic(file_err)
	}
	log.Printf("config file: %s", cfg_file)

	toml_err := toml.Unmarshal([]byte(cfg_file), &cfg)
	if toml_err != nil {
		panic(toml_err)
	}

	log.Printf("config: %v", cfg)

	log.Printf("config.UI: %v", cfg.Ui)

	resolution_cfg := cfg.Ui.Resolution

	log.Printf("x: %v, y: %v", resolution_cfg.X, resolution_cfg.Y)

	ebiten.SetWindowSize(resolution_cfg.X, resolution_cfg.Y)
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
