package main

import (
	"context"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"rounds/pb"
	"sync"

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
	Coords
	ID     string
	Image  *ebiten.Image
	Events chan *pb.ClientEvent
}

type OtherPlayer struct {
	Coords
	ID    string
	Image *ebiten.Image
}

func (p *OtherPlayer) Draw(screen *ebiten.Image) {
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

func (p *Player) OnKeysPressed(keys []ebiten.Key) {
	speed := float64(2)
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		speed *= 1.5
	}

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

	p.Events <- &pb.ClientEvent{
		Event: &pb.ClientEvent_Move{
			Move: &pb.Move{
				PlayerID: p.ID,
				X:        p.X,
				Y:        p.Y,
			},
		},
	}
}

func NewPlayer() *Player {
	image := ebiten.NewImage(16, 16)
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.White)
	return &Player{
		ID:     ksuid.New().String(),
		Image:  image,
		Coords: Coords{32, 32},
		Events: make(chan *pb.ClientEvent, 1024),
	}
}

func NewOtherPlayer(X, Y float64) *OtherPlayer {
	image := ebiten.NewImage(16, 16)
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.RGBA{
		255,
		255,
		0,
		255,
	})
	return &OtherPlayer{
		ID:     ksuid.New().String(),
		Image:  image,
		Coords: Coords{X, Y},
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(p.Image, options)
}

type Game struct {
	drawables []Drawable
	player    *Player
	config    *Config

	// handled off thread. should probably be entities
	otherPlayers map[string]*OtherPlayer
	mu           sync.Mutex
}

func (g *Game) TryAddOtherPlayer(ID string, o *OtherPlayer) {
	// TODO: Need AddPlayer and RemovePlayer ServerEvent so not modifying every
	// frame.
	g.mu.Lock()
	g.otherPlayers[ID] = o
	g.mu.Unlock()
}

func (g *Game) Update() error {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	g.player.OnKeysPressed(inpututil.AppendPressedKeys(keys))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "foobar")

	g.player.Draw(screen)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}
	x, y := ebiten.CursorPosition()
	ebitenutil.DrawLine(screen, g.player.X+8, g.player.Y+8, float64(x), float64(y), color.RGBA{255, 0, 0, 255})

	// lol
	g.mu.Lock()
	log.Println(len(g.otherPlayers))
	for _, otherPlayer := range g.otherPlayers {
		otherPlayer.Draw(screen)
	}
	g.mu.Unlock()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func main() {
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

	// TODO: spin up server if it's not spun up yet

	player := NewPlayer()

	// yolo testing for now
	// beacon player location every 10ms. should do this roughly on demand but fuck yeah
	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "ws://localhost:4242", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close(websocket.StatusInternalError, "")

	game := &Game{
		player:       player,
		drawables:    []Drawable{},
		otherPlayers: make(map[string]*OtherPlayer),
	}

	// TODO: read incoming messages
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
			log.Println(serverEvent.String())

			switch serverEvent.Event.(type) {
			case *pb.ServerEvent_Move:
				move := serverEvent.GetMove()
				if move.PlayerID == player.ID {
					continue
				}
				game.TryAddOtherPlayer(move.PlayerID, NewOtherPlayer(move.X, move.Y))
			}
		}
	}()

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
