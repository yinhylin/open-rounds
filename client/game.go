package client

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"rounds/pb"
	"rounds/world"
	"strings"
	"time"

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
	state           *world.StateBuffer
	playerID        string
	serverEvents    chan *pb.ServerEvent
	clientEvents    chan *pb.ClientEvent
	serverTick      int64
	previousIntents map[pb.Intents_Intent]struct{}
}

func NewGame(assets *Assets) *Game {
	playerID := ksuid.New().String()
	log.Println("you are", playerID)

	clientEvents := make(chan *pb.ClientEvent, 1024)
	clientEvents <- &pb.ClientEvent{
		Id:    playerID,
		Event: &pb.ClientEvent_Connect{},
	}
	clientEvents <- &pb.ClientEvent{
		Id:    playerID,
		Event: &pb.ClientEvent_RequestState{},
	}

	return &Game{
		Assets:          assets,
		state:           world.NewStateBuffer(20),
		playerID:        playerID,
		serverEvents:    make(chan *pb.ServerEvent, 1024),
		serverTick:      world.NilTick,
		clientEvents:    clientEvents,
		previousIntents: make(map[pb.Intents_Intent]struct{}),
	}
}

func (g *Game) requestState() {
	if g.state.CurrentTick() == world.NilTick {
		return
	}
	g.state.Clear()
	g.clientEvents <- &pb.ClientEvent{
		Id:    g.playerID,
		Tick:  g.state.CurrentTick(),
		Event: &pb.ClientEvent_RequestState{},
	}
}

// handleServerEvents drains and applies server events every tick.
func (g *Game) handleServerEvents() error {
	for len(g.serverEvents) > 0 {
		var err error
		select {
		case event := <-g.serverEvents:
			switch event.Event.(type) {
			case *pb.ServerEvent_AddEntity:
				err = g.state.AddEntity(&world.AddEntity{
					Tick: event.Tick,
					ID:   event.GetAddEntity().Entity.Id,
				})

			case *pb.ServerEvent_RemoveEntity:
				err = g.state.RemoveEntity(&world.RemoveEntity{
					Tick: event.Tick,
					ID:   event.GetRemoveEntity().Id,
				})

			case *pb.ServerEvent_EntityEvents:
				msg := event.GetEntityEvents()
				if msg.Id == g.playerID {
					continue
				}
				err = g.state.ApplyIntents(&world.IntentsUpdate{
					Tick:    event.Tick,
					ID:      msg.Id,
					Intents: world.IntentsFromProto(msg.Intents),
				})

			case *pb.ServerEvent_State:
				g.serverTick = event.Tick
				g.state = world.StateBufferFromProto(event.GetState())
				// Simulate next 5 states.
				for i := 0; i < 5; i++ {
					g.state.Next()
				}

			case *pb.ServerEvent_ServerTick:
				g.serverTick = event.Tick
			}
		default:
			return errors.New("should never block")

		}

		if err != nil {
			log.Println(err)
			g.requestState()
		}
	}
	return nil
}

func (g *Game) Update() error {
	now := time.Now()
	defer func() {
		timeTaken := time.Now().Sub(now)
		if timeTaken > time.Millisecond {
			log.Println("long update", time.Now().Sub(now))
		}
	}()

	if err := g.handleServerEvents(); err != nil {
		return err
	}

	if g.state.CurrentTick() == world.NilTick || g.serverTick == world.NilTick {
		return nil
	}

	// Drop a frame if we're too far ahead.
	if g.state.CurrentTick()-g.serverTick > 5 {
		// TODO: Disconnect if this happens too many times in a row without a
		// real frame. Server is dead.
		return nil
	}

	if g.state.CurrentTick() < g.serverTick && g.state.CurrentTick() != world.NilTick {
		// 10 frames behind? Re-request entire server state.
		if math.Abs(float64(g.state.CurrentTick()-g.serverTick)) > 10 {
			log.Println("requesting server state. current tick", g.state.CurrentTick(), "server tick", g.serverTick, "difference:", g.serverTick-g.state.CurrentTick())
			g.requestState()
			return nil
		}

		// 5 frames behind? Skip until we catch up.
		for math.Abs(float64(g.state.CurrentTick()-g.serverTick)) > 5 {
			log.Println("skipping frame. current tick", g.state.CurrentTick(), "server tick", g.serverTick, "difference:", g.serverTick-g.state.CurrentTick())
			g.state.Next()
		}
	}
	g.handleKeysPressed()
	g.state.Next()
	return nil
}

func (g *Game) debugString() string {
	return strings.Join([]string{
		fmt.Sprintf("TPS: %0.02f, FPS: %0.02f", ebiten.CurrentTPS(), ebiten.CurrentFPS()),
		fmt.Sprintf(" T: %d", g.state.CurrentTick()),
		fmt.Sprintf("ST: %d", g.serverTick),
		fmt.Sprintf("DT: %d", g.serverTick-g.state.CurrentTick()),
	}, "\n")
}

func (g *Game) Draw(screen *ebiten.Image) {
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

	screen.Fill(color.RGBA{
		164,
		178,
		191,
		255,
	})

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

	ebitenutil.DebugPrint(screen, g.debugString())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) handleKeysPressed() {
	intents := make(map[pb.Intents_Intent]struct{})
	for _, key := range inpututil.AppendPressedKeys(nil) {
		switch key {
		case ebiten.KeyA:
			intents[pb.Intents_MOVE_LEFT] = struct{}{}
		case ebiten.KeyD:
			intents[pb.Intents_MOVE_RIGHT] = struct{}{}
		case ebiten.KeyW:
			intents[pb.Intents_MOVE_UP] = struct{}{}
		case ebiten.KeyS:
			intents[pb.Intents_MOVE_DOWN] = struct{}{}
		}
	}

	if world.IntentsEqual(g.previousIntents, intents) {
		// Edge triggered.
		return
	}
	g.previousIntents = intents

	tick := g.state.CurrentTick()
	g.state.ApplyIntents(&world.IntentsUpdate{
		ID:      g.playerID,
		Intents: intents,
		Tick:    tick,
	})
	g.clientEvents <- &pb.ClientEvent{
		Id: g.playerID,
		Event: &pb.ClientEvent_Intents{
			Intents: world.IntentsToProto(intents),
		},
		Tick: tick,
	}
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
