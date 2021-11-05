package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Coords struct {
	X, Y float64
}

func (c *Coords) Coordinates() Coords {
	return *c
}

type Drawable interface {
	Coordinates() Coords
	Draw(screen *ebiten.Image)
}

type Player struct {
	Coords
	Image *ebiten.Image
}

func (p *Player) OnKeysPressed(keys []ebiten.Key) {
	speed := float64(2)
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		speed *= 1.5
	}

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
}

func NewPlayer() *Player {
	image := ebiten.NewImage(16, 16)
	ebitenutil.DrawRect(image, 0, 0, 16, 16, color.White)
	return &Player{
		Image:  image,
		Coords: Coords{32, 32},
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(p.Image, options)
}

type Game struct {
	drawables []Drawable
	player    *Player
}

func (g *Game) Update() error {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	g.player.OnKeysPressed(inpututil.AppendPressedKeys(keys))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "foobar")

	g.player.Draw(screen)
	for _, drawable := range g.drawables {
		drawable.Draw(screen)
	}
	x, y := ebiten.CursorPosition()
	ebitenutil.DrawLine(screen, g.player.X+8, g.player.Y+8, float64(x), float64(y), color.RGBA{255, 0, 0, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
}

func main() {
	ebiten.SetWindowSize(640*2, 360*2)
	ebiten.SetWindowTitle("Open ROUNDS")
	if err := ebiten.RunGame(&Game{player: NewPlayer(), drawables: []Drawable{}}); err != nil {
		log.Fatal(err)
	}
}
