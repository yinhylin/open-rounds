package client

import (
	"fmt"
	"math"
	"rounds/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func RenderPlayer(screen *ebiten.Image, image *ebiten.Image, p *world.Player) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Scale(0.1, 0.1)
	opt.GeoM.Translate(p.Coords.X, p.Coords.Y)
	opt.Filter = ebiten.FilterLinear
	_, playerHeight := image.Size()
	playerHeight /= 10
	screen.DrawImage(image, opt)
	debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", p.ID, p.Coords.X, p.Coords.Y)
	ebitenutil.DebugPrintAt(screen, debugString, int(p.Coords.X), int(p.Coords.Y)+playerHeight)
}

func RenderBullet(screen *ebiten.Image, a *Assets, b *world.Bullet) {
	image := a.Image("bullet")
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(32, 32)
	opt.GeoM.Translate(b.Coords.X, b.Coords.Y)
	opt.Filter = ebiten.FilterLinear
	screen.DrawImage(image, opt)
}

func RenderGun(screen *ebiten.Image, a *Assets, coords world.Vector, angle float64) {
	image := a.Image("pistol")
	opt := &ebiten.DrawImageOptions{}
	scale := 0.1
	x := float64(0)
	if math.Abs(angle) > math.Pi/2 {
		scale *= -1
		angle *= -1
		angle += math.Pi
		x += 64
	}
	width, height := image.Size()
	opt.GeoM.Translate(float64(-width/2), float64(-height/2))
	opt.GeoM.Rotate(angle)
	opt.GeoM.Scale(scale, 0.1)
	opt.GeoM.Translate(coords.X+x, coords.Y+32)
	opt.Filter = ebiten.FilterLinear
	screen.DrawImage(image, opt)
}
