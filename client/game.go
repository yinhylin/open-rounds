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
		state:           nil,
		playerID:        playerID,
		serverEvents:    make(chan *pb.ServerEvent, 1024),
		serverTick:      world.NilTick,
		clientEvents:    clientEvents,
		inputDelay:      5,
		previousIntents: make(map[pb.Intents_Intent]struct{}),
	}
}

func (g *Game) requestState() {
	if g.state == nil {
		return
	}
	g.clientEvents <- &pb.ClientEvent{
		Id:    g.playerID,
		Tick:  g.state.CurrentTick(),
		Event: &pb.ClientEvent_RequestState{},
	}
	g.state = nil
}

// handleServerEvents drains and applies server events every tick.
func (g *Game) handleServerEvents() error {
	requestState := false
	for len(g.serverEvents) > 0 {
		select {
		case event := <-g.serverEvents:
			g.serverTick = int64(math.Max(float64(g.serverTick), float64(event.Tick)))
			if player := event.GetPlayer(); player != nil {
				if g.state == nil || player.Id == g.playerID {
					continue
				}
				if err := g.state.OnEvent(event); err != nil {
					log.Println(err)
					requestState = true
				}
				continue
			}

			switch event.Event.(type) {
			case *pb.ServerEvent_State:
				g.state = world.StateBufferFromProto(event.GetState())
				for i := 0; i <= futureStates; i++ {
					g.state.Next()
				}
			}

		default:
			return errors.New("should never block")

		}
	}
	if requestState {
		g.requestState()
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

	if g.state == nil || g.serverTick == world.NilTick {
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
	if g.state == nil {
		screen.Fill(color.RGBA{
			0,
			0,
			0,
			255,
		})
		ebitenutil.DebugPrint(screen, "connecting...")
		return
	}

	g.state.Map().ForEach(func(x, y int64, tile world.Tile) {
		image := g.Image(tile.Image)
		options := &ebiten.DrawImageOptions{}
		options.GeoM.Translate(float64(x*32), float64(y*32))
		screen.DrawImage(image, options)
	})

	g.state.ForEachPlayer(func(ID string, p *world.Player) {
		image := g.Image("zany")
		if ID == g.playerID {
			image = g.Image("cowboy")
		}
		RenderPlayer(screen, image, p)
		RenderGun(screen, g.Assets, p.Coords, p.Angle)
	})

	g.state.ForEachBullet(func(bullet *world.Bullet) {
		RenderBullet(screen, g.Assets, bullet)
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

		e := g.state.Current().Players[g.playerID]
		cX, cY := ebiten.CursorPosition()
		angle := math.Atan2(e.Coords.Y+16-float64(cY), e.Coords.X+16-float64(cX))

		shoot := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
		// TODO: Lower threshold with lerping.
		if shoot || math.Abs(g.previousAngle-angle) > math.Pi/32 {
			g.previousAngle = angle
			g.processClientEvent(&pb.ClientEvent{
				Id: g.playerID,
				Event: &pb.ClientEvent_Angle{
					Angle: &pb.Angle{
						Angle: angle,
					},
				},
				Tick: delayedTick,
			})
		}

		if shoot {
			intents[pb.Intents_SHOOT] = struct{}{}
		}
	}

	if world.IntentsEqual(g.previousIntents, intents) {
		// Edge triggered.
		return
	}
	g.previousIntents = intents
	g.processClientEvent(&pb.ClientEvent{
		Id: g.playerID,
		Event: &pb.ClientEvent_Intents{
			Intents: world.IntentsToProto(intents),
		},
		Tick: delayedTick,
	})
}

func (g *Game) processClientEvent(event *pb.ClientEvent) {
	if serverEvent := world.ClientEventToServerEvent(world.NilTick, event); serverEvent != nil {
		g.state.OnEvent(serverEvent)
	}
	g.clientEvents <- event
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
