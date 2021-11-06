package client

import (
	"context"
	"errors"
	"image/color"
	"io/ioutil"
	"log"
	"rounds/object"
	"rounds/pb"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Drawable interface {
	Coordinates() object.Coords
	Draw(screen *ebiten.Image)
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

type Game struct {
	drawables []Drawable
	player    *LocalPlayer
	config    *Config

	serverEvents chan *pb.ServerEvent
	otherPlayers map[string]*Player
}

func NewGame(player *LocalPlayer) *Game {
	return &Game{
		player:       player,
		drawables:    []Drawable{},
		otherPlayers: make(map[string]*Player),
		serverEvents: make(chan *pb.ServerEvent, 1024),
	}
}

// handleServerEvents drains and applies server events every tick.
func (g *Game) handleServerEvents() error {
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
				if playerID == g.player.ID {
					// TODO: Could not send to the player /shruggie
					continue
				}
				g.otherPlayers[event.PlayerId].Coords = object.Coords{X: move.X, Y: move.Y}
			}
		default:
			return errors.New("should never block")
		}
	}
	return nil
}

func (g *Game) Update() error {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	g.player.OnKeysPressed(inpututil.AppendPressedKeys(keys))
	return g.handleServerEvents()
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{
		109,
		194,
		202,
		255,
	})
	// Draw things (bullets? walls?)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}

	// Draw the player.
	g.player.Draw(screen)

	// Draw the other players with name tags.
	for _, otherPlayer := range g.otherPlayers {
		ebitenutil.DebugPrintAt(screen, otherPlayer.ID, int(otherPlayer.X)-(len(otherPlayer.ID)*5/2), int(otherPlayer.Y)-16)
		otherPlayer.Draw(screen)
	}

	// Draw a line to the cursor.
	x, y := ebiten.CursorPosition()
	ebitenutil.DrawLine(screen, g.player.X+8, g.player.Y+8, float64(x), float64(y), color.RGBA{255, 0, 0, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

// ReadMessages reads the server messages so the game can update accordingly.
func (g *Game) ReadMessages(ctx context.Context, c *websocket.Conn) {
	for {
		messageType, reader, err := c.Reader(ctx)
		if messageType != websocket.MessageBinary {
			log.Fatal(messageType)
		}
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
		g.serverEvents <- &serverEvent
	}
}
