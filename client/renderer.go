package client

import (
	"fmt"
	"math"
	"time"

	"github.com/sailormoon/open-rounds/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type RenderData struct {
	lastPlayerDrawCoords world.Vector
	lastGunDrawCoords    world.Vector
	prevAngle            float64
	targetAngle          float64
}

type Renderer struct {
	renderData   map[string]*RenderData
	lastDrawTime time.Time
	lerpFactor   float64
}

func NewRenderer(lerpFactor float64) *Renderer {
	return &Renderer{
		renderData: make(map[string]*RenderData),
		lerpFactor: lerpFactor,
	}
}

func (r *Renderer) RenderPlayer(screen *ebiten.Image, playerImage *ebiten.Image, gunImage *ebiten.Image, p *world.Player) {
	opt := &ebiten.DrawImageOptions{}
	x, y := p.Coords.X, p.Coords.Y
	if renderData, ok := r.renderData[p.ID]; ok {
		if time.Since(r.lastDrawTime).Seconds() <= r.lerpFactor {
			lerpDelta := time.Since(r.lastDrawTime).Seconds() / r.lerpFactor
			x = Lerp(renderData.lastPlayerDrawCoords.X, x, lerpDelta)
			y = Lerp(renderData.lastPlayerDrawCoords.Y, y, lerpDelta)
		}
	} else {
		r.renderData[p.ID] = &RenderData{
			lastPlayerDrawCoords: world.Vector{X: x, Y: y},
			prevAngle:            p.Angle,
		}
	}
	opt.GeoM.Translate(x, y)
	opt.Filter = ebiten.FilterLinear
	_, playerHeight := playerImage.Size()
	r.renderData[p.ID].lastPlayerDrawCoords = world.Vector{X: x, Y: y}
	screen.DrawImage(playerImage, opt)
	RenderGun(screen, gunImage, r.renderData[p.ID].lastPlayerDrawCoords, p.Angle)
	debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)", p.ID, p.Coords.X, p.Coords.Y)
	ebitenutil.DebugPrintAt(screen, debugString, int(x), int(y)+playerHeight)
}

func (r *Renderer) RenderBullet(screen *ebiten.Image, a *Assets, b *world.Bullet) {
	image := a.Image("bullet")
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(16, 16)
	opt.GeoM.Translate(b.Coords.X, b.Coords.Y)
	opt.Filter = ebiten.FilterLinear
	screen.DrawImage(image, opt)
}

func RenderGun(screen *ebiten.Image, image *ebiten.Image, coords world.Vector, angle float64) {
	opt := &ebiten.DrawImageOptions{}
	scale := 0.1
	x := float64(0)
	if math.Abs(angle) > math.Pi/2 {
		scale *= -1
		angle *= -1
		angle += math.Pi
		x += 32
	}
	width, height := image.Size()
	opt.GeoM.Translate(float64(-width/2), float64(-height/2))
	opt.GeoM.Rotate(angle)
	opt.GeoM.Scale(scale, 0.1)
	opt.GeoM.Translate(coords.X+x, coords.Y+16)
	opt.Filter = ebiten.FilterLinear
	screen.DrawImage(image, opt)
}

func Lerp(start, end, t float64) float64 {
	return start*(1.0-t) + end*t
}
