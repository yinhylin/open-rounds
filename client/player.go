package client

import (
	"fmt"
	"rounds/world"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/segmentio/ksuid"
)

type Player struct {
	world.Entity
	ID    string
	Image *ebiten.Image
}

type LocalPlayer struct {
	Player
	keys []ebiten.Key
}

func (p *Player) debugString() string {
	return strings.Join([]string{
		fmt.Sprintf("ID:  %s", p.ID),
		fmt.Sprintf("x,y: %0.0f,%0.0f", p.X, p.Y),
	}, "\n")
}

func (p *Player) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(p.Image, options)
	ebitenutil.DebugPrintAt(screen, p.debugString(), int(p.X), int(p.Y)+16)
}

func NewLocalPlayer(image *ebiten.Image) *LocalPlayer {
	player := &LocalPlayer{
		Player: Player{
			ID:    ksuid.New().String(),
			Image: image,
		},
	}
	return player
}

func NewOtherPlayer(ID string, X, Y float64, image *ebiten.Image) *Player {
	return &Player{
		ID:    ID,
		Image: image,
		Entity: world.Entity{
			Coords: world.Coords{X: X, Y: Y},
		},
	}
}
