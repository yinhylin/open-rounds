package client

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"
	"rounds/pb"
	"rounds/world"
	"strings"
	"time"

	"github.com/ebiten/emoji"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/segmentio/ksuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wspb"
)

const (
	futureStates = 5
)

type Game struct {
	*Assets
	state           *world.StateBuffer
	playerID        string
	serverEvents    chan *pb.ServerEvent
	clientEvents    chan *pb.ClientEvent
	serverTick      int64
	inputDelay      int64
	previousAngle   float64
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
		inputDelay:      3,
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
			g.serverTick = int64(math.Max(float64(g.serverTick), float64(event.ServerTick)))
			switch event.Event.(type) {
			case *pb.ServerEvent_AddEntity:
				if g.state.Current() == nil {
					continue
				}
				err = g.state.AddEntity(&world.AddEntity{
					Tick: event.Tick,
					ID:   event.GetAddEntity().Entity.Id,
				})

			case *pb.ServerEvent_RemoveEntity:
				if g.state.Current() == nil {
					continue
				}
				err = g.state.RemoveEntity(&world.RemoveEntity{
					Tick: event.Tick,
					ID:   event.GetRemoveEntity().Id,
				})

			case *pb.ServerEvent_EntityEvents:
				if g.state.CurrentTick() == world.NilTick {
					continue
				}
				msg := event.GetEntityEvents()
				if msg.Id == g.playerID {
					// TODO: Store a rolling buffer of input delay and ease instead of updating immediately.
					difference := g.state.CurrentTick() - event.Tick
					if difference > 2 {
						g.inputDelay++
					}
					if difference < 1 {
						g.inputDelay--
					}
					continue
				}
				err = g.state.ApplyIntents(&world.IntentsUpdate{
					Tick:    event.Tick,
					ID:      msg.Id,
					Intents: world.IntentsFromProto(msg.Intents),
				})

			case *pb.ServerEvent_State:
				g.state = world.StateBufferFromProto(event.GetState())
				// Simulate next N states.
				for i := 0; i <= futureStates; i++ {
					g.state.Next()
				}

			case *pb.ServerEvent_EntityAngle:
				if g.state.Current() == nil {
					continue
				}
				msg := event.GetEntityAngle()
				if msg.Id == g.playerID {
					continue
				}
				g.state.ApplyAngle(&world.AngleUpdate{
					Tick:  g.state.CurrentTick(),
					ID:    msg.Id,
					Angle: msg.Angle,
				})
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
		if time.Since(now) > time.Millisecond {
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
	if g.state.CurrentTick()-g.serverTick > futureStates {
		// TODO: Disconnect if this happens too many times in a row without a
		// real frame. Server is dead.
		return nil
	}

	if g.state.CurrentTick() < g.serverTick && g.state.CurrentTick() != world.NilTick {
		// 10 frames behind? Re-request entire server state.
		if g.serverTick-g.state.CurrentTick() > 10 {
			log.Println("requesting server state. current tick", g.state.CurrentTick(), "server tick", g.serverTick, "difference:", g.serverTick-g.state.CurrentTick())
			g.requestState()
			return nil

		}

		for g.serverTick-g.state.CurrentTick() > 5 {
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
		fmt.Sprintf("ID: %d", g.inputDelay),
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
		image := emoji.Image("🥴")
		if ID == g.playerID {
			image = emoji.Image("🤠")
		}

		options := &ebiten.DrawImageOptions{}
		options.GeoM.Scale(0.5, 0.5)
		options.GeoM.Translate(e.Coords.X, e.Coords.Y)
		options.Filter = ebiten.FilterLinear
		_, height := image.Size()
		height /= 2
		screen.DrawImage(image, options)

		gun := emoji.Image("🔫")
		options = &ebiten.DrawImageOptions{}
		angle := e.Angle
		scale := 0.4
		x := float64(0)
		if math.Abs(angle) > math.Pi/2 {
			scale *= -1
			angle *= -1
			angle += math.Pi
			x += 64
		}
		options.GeoM.Translate(-64, -64)
		options.GeoM.Rotate(angle)
		options.GeoM.Scale(scale, 0.4)
		options.GeoM.Translate(e.Coords.X+x, e.Coords.Y+48)
		options.Filter = ebiten.FilterLinear
		screen.DrawImage(gun, options)

		debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", ID, e.Coords.X, e.Coords.Y)
		ebitenutil.DebugPrintAt(screen, debugString, int(e.Coords.X), int(e.Coords.Y)+height)
	})

	ebitenutil.DebugPrint(screen, g.debugString())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) handleKeysPressed() {
	intents := make(map[pb.Intents_Intent]struct{})
	if ebiten.IsFocused() {
		for _, key := range inpututil.AppendPressedKeys(nil) {
			switch key {
			case ebiten.KeyA:
				intents[pb.Intents_MOVE_LEFT] = struct{}{}
			case ebiten.KeyD:
				intents[pb.Intents_MOVE_RIGHT] = struct{}{}
			case ebiten.KeyW, ebiten.KeySpace:
				intents[pb.Intents_MOVE_UP] = struct{}{}
			case ebiten.KeyS:
				intents[pb.Intents_MOVE_DOWN] = struct{}{}
			}
		}

		e := g.state.Current().Entities[g.playerID]
		cX, cY := ebiten.CursorPosition()
		angle := math.Atan2(e.Coords.Y-float64(cY), e.Coords.X-float64(cX))
		g.state.ApplyAngle(&world.AngleUpdate{
			Tick:  g.state.CurrentTick(),
			ID:    g.playerID,
			Angle: angle,
		})

		// TODO: Lower threshold with lerping.
		if math.Abs(g.previousAngle-angle) > math.Pi/32 {
			g.previousAngle = angle
			g.clientEvents <- &pb.ClientEvent{
				Id: g.playerID,
				Event: &pb.ClientEvent_Angle{
					Angle: &pb.Angle{
						Angle: angle,
					},
				},
				Tick: g.state.CurrentTick() + 3,
			}
		}
	}

	if world.IntentsEqual(g.previousIntents, intents) {
		// Edge triggered.
		return
	}
	g.previousIntents = intents

	tick := g.state.CurrentTick() + g.inputDelay
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
		var serverEvent pb.ServerEvent
		if err := wspb.Read(ctx, c, &serverEvent); err != nil {
			log.Fatal(err)
		}
		g.serverEvents <- &serverEvent
	}
}

// WriteMessages takes the game acitons and sends it to the server.
func (g *Game) WriteMessages(ctx context.Context, c *websocket.Conn) {
	for event := range g.clientEvents {
		if err := wspb.Write(ctx, c, event); err != nil {
			log.Fatal(err)
		}
	}
	c.Close(websocket.StatusInternalError, "reader closed")
}
