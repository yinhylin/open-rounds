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
	m               *world.Map
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
		m:               assets.Map("basic"),
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
			if g.state.Current() == nil {
				if _, ok := event.Event.(*pb.ServerEvent_State); !ok {
					continue
				}
			}

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
				msg := event.GetEntityAngle()
				if msg.Id == g.playerID {
					continue
				}
				g.state.ApplyAngle(&world.AngleUpdate{
					Tick:  event.Tick,
					ID:    msg.Id,
					Angle: msg.Angle,
				})

			case *pb.ServerEvent_EntityShoot:
				msg := event.GetEntityShoot()
				if msg.SourceId == g.playerID {
					continue
				}
				g.state.AddBullet(&world.AddBullet{
					Tick:   event.Tick,
					Source: msg.SourceId,
					ID:     msg.Id,
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
		if g.serverTick-g.state.CurrentTick() > 20 {
			log.Println("requesting server state. current tick", g.state.CurrentTick(), "server tick", g.serverTick, "difference:", g.serverTick-g.state.CurrentTick())
			g.requestState()
			return nil

		}

		for g.serverTick-g.state.CurrentTick() > futureStates {
			log.Println("skipping frame. current tick", g.state.CurrentTick(), "server tick", g.serverTick, "difference:", g.serverTick-g.state.CurrentTick())
			g.state.Next()
		}
	}
	g.handleInput()
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

	g.m.ForEach(func(x, y int, tile world.Tile) {
		image := g.Image(tile.Image)
		options := &ebiten.DrawImageOptions{}
		options.GeoM.Translate(float64(x*32), float64(y*32))
		screen.DrawImage(image, options)
	})

	g.state.ForEachEntity(func(ID string, e *world.Entity) {
		// lol
		image := emoji.Image("🥴")
		if ID == g.playerID {
			image = emoji.Image("🤠")
		}

		bodyOptions := &ebiten.DrawImageOptions{}
		bodyOptions.GeoM.Scale(0.5, 0.5)
		bodyOptions.GeoM.Translate(e.Coords.X, e.Coords.Y)
		bodyOptions.Filter = ebiten.FilterLinear
		_, height := image.Size()
		height /= 2

		gun := emoji.Image("🔫")
		gunOptions := &ebiten.DrawImageOptions{}
		angle := e.Angle
		scale := 0.4
		x := float64(0)
		if math.Abs(angle) > math.Pi/2 {
			scale *= -1
			angle *= -1
			angle += math.Pi
			x += 64
		}
		gunOptions.GeoM.Translate(-64, -64)
		gunOptions.GeoM.Rotate(angle)
		gunOptions.GeoM.Scale(scale, 0.4)
		gunOptions.GeoM.Translate(e.Coords.X+x, e.Coords.Y+48)
		gunOptions.Filter = ebiten.FilterLinear

		screen.DrawImage(image, bodyOptions)
		screen.DrawImage(gun, gunOptions)

		debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", ID, e.Coords.X, e.Coords.Y)
		ebitenutil.DebugPrintAt(screen, debugString, int(e.Coords.X), int(e.Coords.Y)+height)
	})

	g.state.ForEachBullet(func(ID string, e *world.Bullet) {
		image := emoji.Image("🔴")
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Scale(0.1, 0.1)
		opt.GeoM.Translate(32, 32)
		opt.GeoM.Translate(e.Coords.X, e.Coords.Y)
		screen.DrawImage(image, opt)
	})

	ebitenutil.DebugPrint(screen, g.debugString())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1280, 720
}

func (g *Game) handleInput() {
	intents := make(map[pb.Intents_Intent]struct{})
	delayedTick := g.state.CurrentTick() + g.inputDelay
	if ebiten.IsFocused() {
		for _, key := range inpututil.AppendPressedKeys(nil) {
			switch key {
			case ebiten.KeyA:
				intents[pb.Intents_MOVE_LEFT] = struct{}{}
			case ebiten.KeyD:
				intents[pb.Intents_MOVE_RIGHT] = struct{}{}
			case ebiten.KeyW, ebiten.KeySpace:
				intents[pb.Intents_JUMP] = struct{}{}
			}
		}

		e := g.state.Current().Entities[g.playerID]
		cX, cY := ebiten.CursorPosition()
		angle := math.Atan2(e.Coords.Y-float64(cY), e.Coords.X-float64(cX))

		shoot := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
		// TODO: Lower threshold with lerping.
		if shoot || math.Abs(g.previousAngle-angle) > math.Pi/32 {
			g.previousAngle = angle
			g.state.ApplyAngle(&world.AngleUpdate{
				Tick:  delayedTick,
				ID:    g.playerID,
				Angle: angle,
			})
			g.clientEvents <- &pb.ClientEvent{
				Id: g.playerID,
				Event: &pb.ClientEvent_Angle{
					Angle: &pb.Angle{
						Angle: angle,
					},
				},
				Tick: delayedTick,
			}
		}

		if shoot {
			shotID := ksuid.New().String()
			g.state.AddBullet(&world.AddBullet{
				Source: g.playerID,
				ID:     shotID,
				Tick:   delayedTick,
			})
			g.clientEvents <- &pb.ClientEvent{
				Id: g.playerID,
				Event: &pb.ClientEvent_Shoot{
					Shoot: &pb.Shoot{
						Id: shotID,
					},
				},
				Tick: delayedTick,
			}
		}
	}

	if world.IntentsEqual(g.previousIntents, intents) {
		// Edge triggered.
		return
	}
	g.previousIntents = intents

	g.state.ApplyIntents(&world.IntentsUpdate{
		ID:      g.playerID,
		Intents: intents,
		Tick:    delayedTick,
	})
	g.clientEvents <- &pb.ClientEvent{
		Id: g.playerID,
		Event: &pb.ClientEvent_Intents{
			Intents: world.IntentsToProto(intents),
		},
		Tick: delayedTick,
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
