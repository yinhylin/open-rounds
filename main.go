package main

import (
	"context"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"rounds/object"
	"rounds/pb"
	"rounds/server"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/pelletier/go-toml/v2"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Coords struct {
	X, Y float64
}

func (c *Coords) Coordinates() Coords {
	return *c
}

type Drawable interface {
	Coordinates() Coords
	Draw(screen *ebiten.Image)
}

type Player struct {
	object.Entity
	ID    string
	Image *ebiten.Image
}

type LocalPlayer struct {
	Player
	Events chan *pb.ClientEvent
}

func (p *Player) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(p.Image, options)
}

type PlayerConfig struct {
	speed      float32
	jumpHeight float32
}

type GameConfig struct {
	Bar string
}

type ResolutionConfig struct {
	X, Y int
}

type UIConfig struct {
	Resolution ResolutionConfig
}

type Config struct {
	Player PlayerConfig
	Ui     UIConfig
	Game   GameConfig
}

func (p *LocalPlayer) OnKeysPressed(keys []ebiten.Key) {
	speed := float64(2)
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		speed *= 1.5
	}

	oldCoords := p.Coords

	for _, key := range keys {
		switch key {
		case ebiten.KeyA:
			p.Coords.X -= speed
		case ebiten.KeyD:
			p.Coords.X += speed
		case ebiten.KeyW:
			p.Coords.Y -= speed
		case ebiten.KeyS:
			p.Coords.Y += speed
		case ebiten.KeyQ, ebiten.KeyEscape:
			log.Fatal("quit")
		}
	}

	if oldCoords != p.Coords {
		p.Events <- &pb.ClientEvent{
			PlayerUuid: p.ID,
			Event: &pb.ClientEvent_Move{
				Move: &pb.Move{
					X: p.X,
					Y: p.Y,
				},
			},
		}
	}
}

func NewLocalPlayer() *LocalPlayer {
	image := ebiten.NewImage(16, 16)
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.White)
	player := &LocalPlayer{
		Player: Player{
			ID:    ksuid.New().String(),
			Image: image,
			Entity: object.Entity{
				Coords: object.Coords{X: 32, Y: 32},
			},
		},
		Events: make(chan *pb.ClientEvent, 1024),
	}
	player.Events <- &pb.ClientEvent{
		PlayerUuid: player.ID,
		Event:      &pb.ClientEvent_Connect{},
	}
	player.Events <- &pb.ClientEvent{
		PlayerUuid: player.ID,
		Event: &pb.ClientEvent_Move{
			Move: &pb.Move{
				X: player.X,
				Y: player.Y,
			},
		},
	}
	return player
}

func NewOtherPlayer(ID string, X, Y float64) *Player {
	image := ebiten.NewImage(16, 16)
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.RGBA{
		255,
		255,
		0,
		255,
	})
	return &Player{
		ID:    ID,
		Image: image,
		Entity: object.Entity{
			Coords: object.Coords{X: X, Y: Y},
		},
	}
}

type Game struct {
	drawables []Drawable
	player    *LocalPlayer
	config    *Config

	serverEvents chan *pb.ServerEvent
	otherPlayers map[string]*Player
}

func (g *Game) Update() error {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	g.player.OnKeysPressed(inpututil.AppendPressedKeys(keys))

	// Drain server events.
	for len(g.serverEvents) > 0 {
		select {
		case event := <-g.serverEvents:
			playerID := event.PlayerId
			if playerID == g.player.ID {
				// TODO: Could not send to the player /shruggie
				continue
			}

			switch event.Event.(type) {
			case *pb.ServerEvent_AddPlayer:
				g.otherPlayers[event.PlayerId] = NewOtherPlayer(event.PlayerId, 32, 32)
			case *pb.ServerEvent_RemovePlayer:
				delete(g.otherPlayers, event.PlayerId)
			case *pb.ServerEvent_Move:
				move := event.GetMove()
				g.otherPlayers[event.PlayerId].Coords = object.Coords{X: move.X, Y: move.Y}
			}
		default:
			log.Println("should never block")
			break
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "foobar")

	g.player.Draw(screen)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}

	for _, otherPlayer := range g.otherPlayers {
		ebitenutil.DebugPrintAt(screen, otherPlayer.ID, int(otherPlayer.X)-(len(otherPlayer.ID)*5/2), int(otherPlayer.Y)-16)
		otherPlayer.Draw(screen)
	}

	x, y := ebiten.CursorPosition()
	ebitenutil.DrawLine(screen, g.player.X+8, g.player.Y+8, float64(x), float64(y), color.RGBA{255, 0, 0, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	// Load Configs
	var cfg Config
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

	player := NewLocalPlayer()
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

	game := &Game{
		player:       player,
		drawables:    []Drawable{},
		otherPlayers: make(map[string]*Player),
		serverEvents: make(chan *pb.ServerEvent, 1024),
	}

	// reader
	go func() {
		for {
			_, reader, err := c.Reader(ctx)
			if err != nil {
				log.Fatal(err)
			}

			b, err := ioutil.ReadAll(reader)
			if err != nil {
				// TODO: close connection / reconnect
				log.Fatal(err)
			}
			if len(b) <= 0 {
				continue
			}

			var serverEvent pb.ServerEvent
			err = proto.Unmarshal(b, &serverEvent)
			if err != nil {
				// TODO: close connection / reconnect
				log.Fatal(err)
			}
			game.serverEvents <- &serverEvent
		}
	}()

	// the writer
	// https://www.youtube.com/watch?v=H-ru2glqXAg
	go func() {
		for event := range player.Events {
			bytes, err := proto.Marshal(event)
			// TODO: handle these more gracefully
			if err != nil {
				log.Fatal(err)
			}
			if err := c.Write(ctx, websocket.MessageBinary, bytes); err != nil {
				log.Fatal(err)
			}
		}
	}()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
