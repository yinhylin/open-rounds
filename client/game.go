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

type Drawable interface {
	Coordinates() world.Coords
	Draw(screen *ebiten.Image)
}

func lerp(v0, v1, t float64) float64 {
	return (1-t)*v0 + t*v1
}

type LerpState struct {
	source    world.Coords
	target    world.Coords
	iteration float64
}

func (l *LerpState) Current() world.Coords {
	return world.Coords{
		X: lerp(l.source.X, l.target.X, l.iteration),
		Y: lerp(l.source.Y, l.target.Y, l.iteration),
	}
}

func (l *LerpState) Advance() {
	l.iteration = math.Min(l.iteration+0.25, 1.00)
}

type Game struct {
	*Assets
	state        *world.World
	drawables    []Drawable
	player       *world.Entity
	playerID     string
	lerps        map[string]*LerpState
	serverEvents chan *pb.ServerEvent
	clientEvents chan *pb.ClientEvent
}

func NewGame(assets *Assets) *Game {
	state := world.NewWorld()
	player := &world.Entity{}
	playerID := ksuid.New().String()
	state.AddEntity(playerID, player)

	clientEvents := make(chan *pb.ClientEvent, 1024)
	clientEvents <- &pb.ClientEvent{
		Id:    playerID,
		Event: &pb.ClientEvent_Connect{},
	}

	return &Game{
		Assets:       assets,
		state:        state,
		player:       player,
		playerID:     playerID,
		drawables:    []Drawable{},
		lerps:        make(map[string]*LerpState),
		serverEvents: make(chan *pb.ServerEvent, 1024),
		clientEvents: clientEvents,
	}
}

// handleServerEvents drains and applies server events every tick.
func (g *Game) handleServerEvents() error {
	for len(g.serverEvents) > 0 {
		select {
		case event := <-g.serverEvents:
			if event.Id == g.playerID {
				continue
			}
			switch event.Event.(type) {
			case *pb.ServerEvent_AddPlayer:
				if event.Id != g.playerID {
					g.state.AddEntity(event.Id, &world.Entity{})
				}

			case *pb.ServerEvent_RemovePlayer:
				g.state.RemoveEntity(event.Id)

			case *pb.ServerEvent_SetPosition:
				position := event.GetSetPosition()
				entity := g.state.Entity(event.Id)
				if entity == nil {
					continue
				}
				target := world.Coords{
					X: position.Position.X,
					Y: position.Position.Y,
				}
				currentLerp := g.lerps[event.Id]
				log.Printf("%+v vs %+v\n", currentLerp, target)
				if currentLerp == nil || currentLerp.target != target {
					source := entity.Coords
					if currentLerp != nil {
						source = currentLerp.Current()
					}
					g.lerps[event.Id] = &LerpState{
						source: source,
						target: target,
					}
				}
				entity.Coords = target

			case *pb.ServerEvent_EntityState:
				state := event.GetEntityState()
				entity := g.state.Entity(event.Id)
				if entity == nil {
					continue
				}
				target := world.Coords{
					X: state.Position.X,
					Y: state.Position.Y,
				}
				currentLerp := g.lerps[event.Id]
				log.Printf("%+v vs %+v\n", currentLerp, target)
				if currentLerp == nil || currentLerp.target != target {
					source := entity.Coords
					if currentLerp != nil {
						source = currentLerp.Current()
					}
					g.lerps[event.Id] = &LerpState{
						source: source,
						target: target,
					}
				}
				entity.Coords = target

				entity.Velocity = world.Vector{
					X: state.Velocity.X,
					Y: state.Velocity.Y,
				}
			}
		default:
			return errors.New("should never block")
		}
	}
	return nil
}

func (g *Game) Update() error {
	oldCoords := g.player.Coords
	if err := g.handleServerEvents(); err != nil {
		return err
	}
	g.handleKeysPressed()
	g.state.Update()
	if g.player.Coords != oldCoords {
		g.clientEvents <- &pb.ClientEvent{
			Id: g.playerID,
			Event: &pb.ClientEvent_SetPosition{
				SetPosition: &pb.SetPosition{
					Position: &pb.Vector{
						X: g.player.X,
						Y: g.player.Y,
					},
				}},
		}
	}
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
	// Draw things (bullets? walls?)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}

	g.state.ForEachEntity(func(ID string, e *world.Entity) {
		// lol
		image := g.Image("enemy")
		if ID == g.playerID {
			image = g.Image("player")
		}

		options := &ebiten.DrawImageOptions{}
		var coords world.Coords
		if lerp := g.lerps[ID]; lerp != nil {
			lerp.Advance()
			coords = lerp.Current()
			options.GeoM.Translate(coords.X, coords.Y)
		} else {
			coords = e.Coords
			options.GeoM.Translate(e.X, e.Y)
		}
		screen.DrawImage(image, options)
		debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", ID, e.X, e.Y)
		ebitenutil.DebugPrintAt(screen, debugString, int(coords.X), int(coords.Y)+16)
	})

	// Draw a line to the cursor.
	x, y := ebiten.CursorPosition()
	ebitenutil.DebugPrint(screen, debugString())
	ebitenutil.DrawLine(screen, g.player.X+8, g.player.Y+8, float64(x), float64(y), color.RGBA{255, 0, 0, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) handleKeysPressed() {
	var actions []*pb.Action
	var keys []ebiten.Key
	for _, key := range inpututil.AppendPressedKeys(keys) {
		switch key {
		case ebiten.KeyA:
			actions = append(actions, &pb.Action{
				Action: pb.Action_MOVE_LEFT,
			})
		case ebiten.KeyD:
			actions = append(actions, &pb.Action{
				Action: pb.Action_MOVE_RIGHT,
			})
		case ebiten.KeyW:
			actions = append(actions, &pb.Action{
				Action: pb.Action_MOVE_UP,
			})
		case ebiten.KeyS:
			actions = append(actions, &pb.Action{
				Action: pb.Action_MOVE_DOWN,
			})
		}
	}

	if actions == nil {
		actions = append(actions, &pb.Action{
			Action: pb.Action_NONE,
		})
	}
	g.clientEvents <- &pb.ClientEvent{
		Id: g.playerID,
		Event: &pb.ClientEvent_Actions{
			Actions: &pb.Actions{
				Actions: actions,
			},
		},
	}
	g.player.OnActions(actions)
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
