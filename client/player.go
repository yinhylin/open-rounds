package client

import (
	"rounds/world"

	"github.com/hajimehoshi/ebiten/v2"
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

func NewLocalPlayer(image *ebiten.Image) *LocalPlayer {
	player := &LocalPlayer{
		Player: Player{
			ID:    ksuid.New().String(),
			Image: image,
		},
	}
	return player
}
