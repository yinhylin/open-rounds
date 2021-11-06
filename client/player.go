package client

import (
	"context"
	"image/color"
	"log"
	"rounds/object"
	"rounds/pb"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

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
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.RGBA{
		218,
		212,
		94,
		255,
	})
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
		208,
		70,
		72,
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

// WriteMessages takes the player input and writes it to the server.
func (p *LocalPlayer) WriteMessages(ctx context.Context, c *websocket.Conn) {
	for event := range p.Events {
		bytes, err := proto.Marshal(event)
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Write(ctx, websocket.MessageBinary, bytes); err != nil {
			log.Fatal(err)
		}
	}
}
