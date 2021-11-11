package client

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"rounds/pb"
	"rounds/world"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

type Updatable interface {
	Update()
}

type Game struct {
	*Assets
	state        *world.StateBuffer
	player       *world.Entity
	playerID     string
	serverEvents chan *pb.ServerEvent
	clientEvents chan *pb.ClientEvent
}

func NewGame(assets *Assets) *Game {
	player := &world.Entity{}
	playerID := ksuid.New().String()

	clientEvents := make(chan *pb.ClientEvent, 1024)
	clientEvents <- &pb.ClientEvent{
		Id:    playerID,
		Event: &pb.ClientEvent_Connect{},
	}

	return &Game{
		Assets:       assets,
		state:        world.NewStateBuffer(8),
		player:       player,
		playerID:     playerID,
		serverEvents: make(chan *pb.ServerEvent, 1024),
		clientEvents: clientEvents,
	}
}

// handleServerEvents drains and applies server events every tick.
func (g *Game) handleServerEvents() error {
	for len(g.serverEvents) > 0 {
		select {
		case event := <-g.serverEvents:
			switch event.Event.(type) {
			case *pb.ServerEvent_AddEntity:
				g.state.AddEntity(&world.AddEntity{
					Tick: event.Tick,
					ID:   event.GetAddEntity().Entity.Id,
				})

			case *pb.ServerEvent_RemoveEntity:
				g.state.RemoveEntity(&world.RemoveEntity{
					Tick: event.Tick,
					ID:   event.GetRemoveEntity().Id,
				})

			case *pb.ServerEvent_EntityEvents:
				msg := event.GetEntityEvents()
				g.state.ApplyActions(&world.ActionsUpdate{
					Tick:    event.Tick,
					ID:      msg.Id,
					Actions: world.ActionsFromProto(msg.Actions),
				})

			case *pb.ServerEvent_States:
				msg := event.GetStates()
				state := &world.State{
					Simulated: false,
					Entities:  make(map[string]world.Entity),
					Tick:      event.Tick,
				}
				for _, entity := range msg.States {
					state.Entities[entity.Id] = *world.EntityFromProto(entity)
				}
				g.state.Add(state)
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
	if g.state.CurrentTick() == -1 {
		return nil
	}
	g.handleKeysPressed()
	g.state.Next()
	return nil
}

func debugString() string {
	return strings.Join([]string{
		fmt.Sprintf("Version: %s, TPS: %0.02f, FPS: %0.02f", strings.TrimSpace(Version), ebiten.CurrentTPS(), ebiten.CurrentFPS()),
	}, "\n")
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{
		164,
		178,
		191,
		255,
	})

	if g.state.Current() == nil {
		screen.Fill(color.RGBA{
			0,
			0,
			0,
			255,
		})
		ebitenutil.DebugPrint(screen, "connecting...")
		return
	}

	serverState := g.state.CurrentServer()

	for _, e := range serverState.Entities {
		ebitenutil.DrawRect(screen, e.Coords.X, e.Coords.Y, 16, 16, color.RGBA{188, 0, 0, 255})
	}

	g.state.ForEachEntity(func(ID string, e *world.Entity) {
		// lol
		image := g.Image("enemy")
		if ID == g.playerID {
			image = g.Image("player")
		}

		options := &ebiten.DrawImageOptions{}
		options.GeoM.Translate(e.Coords.X, e.Coords.Y)
		screen.DrawImage(image, options)
		debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", ID, e.Coords.X, e.Coords.Y)
		ebitenutil.DebugPrintAt(screen, debugString, int(e.Coords.X), int(e.Coords.Y)+16)
	})

	ebitenutil.DebugPrint(screen, debugString())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) handleKeysPressed() {
	actions := make(map[pb.Actions_Event]struct{})
	for _, key := range inpututil.AppendPressedKeys(nil) {
		switch key {
		case ebiten.KeyA:
			actions[pb.Actions_MOVE_LEFT] = struct{}{}
		case ebiten.KeyD:
			actions[pb.Actions_MOVE_RIGHT] = struct{}{}
		case ebiten.KeyW:
			actions[pb.Actions_MOVE_UP] = struct{}{}
		case ebiten.KeyS:
			actions[pb.Actions_MOVE_DOWN] = struct{}{}
		}
	}

	tick := g.state.CurrentTick()
	g.clientEvents <- &pb.ClientEvent{
		Id: g.playerID,
		Event: &pb.ClientEvent_Actions{
			Actions: world.ActionsToProto(actions),
		},
		Tick: tick,
	}

	g.state.ApplySimulatedActions(&world.ActionsUpdate{
		ID:      g.playerID,
		Actions: actions,
		Tick:    tick,
	})
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

// WriteMessages takes the game acitons and sends it to the server.
func (g *Game) WriteMessages(ctx context.Context, c *websocket.Conn) {
	for event := range g.clientEvents {
		bytes, err := proto.Marshal(event)
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Write(ctx, websocket.MessageBinary, bytes); err != nil {
			log.Fatal(err)
		}
	}
}
