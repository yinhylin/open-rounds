package client

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"rounds/object"
	"rounds/pb"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Updatable interface {
	Update()
}

type Drawable interface {
	Coordinates() object.Coords
	Draw(screen *ebiten.Image)
}

type Game struct {
	*Assets
	drawables    []Drawable
	updatables   []Updatable
	player       *LocalPlayer
	serverEvents chan *pb.ServerEvent
	otherPlayers map[string]*Player
}

func NewGame(player *LocalPlayer, assets *Assets) *Game {
	return &Game{
		Assets:       assets,
		player:       player,
		drawables:    []Drawable{player},
		updatables:   []Updatable{player},
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
				g.otherPlayers[event.PlayerId] = NewOtherPlayer(event.PlayerId, 32, 32, g.Image("enemy"))
			case *pb.ServerEvent_RemovePlayer:
				delete(g.otherPlayers, event.PlayerId)
			case *pb.ServerEvent_Move:
				move := event.GetMove()
				g.otherPlayers[event.PlayerId].Coords = object.Coords{X: move.X, Y: move.Y}
			}
		default:
			return errors.New("should never block")
		}
	}
	return nil
}

func (g *Game) Update() error {
	if err := g.handleServerEvents(); err != nil {
		return err
	}
	for _, updatable := range g.updatables {
		updatable.Update()
	}
	return nil
}

func debugString() string {
	return strings.Join([]string{
		fmt.Sprintf("Version: %s", strings.TrimSpace(Version)),
		fmt.Sprintf("TPS:     %0.2f", ebiten.CurrentTPS()),
		fmt.Sprintf("FPS:     %0.2f", ebiten.CurrentFPS()),
	}, "\n")
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{
		164,
		178,
		191,
		255,
	})
	// Draw things (bullets? walls?)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}

	// Draw the other players with name tags.
	for _, otherPlayer := range g.otherPlayers {
		otherPlayer.Draw(screen)
	}

	// Draw a line to the cursor.
	x, y := ebiten.CursorPosition()
	ebitenutil.DebugPrint(screen, debugString())
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
			continue
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
