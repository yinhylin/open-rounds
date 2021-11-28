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
}

func NewRenderer() *Renderer {
	return &Renderer{
		renderData: make(map[string]*RenderData),
	}
}

func (r *Renderer) RenderPlayer(screen *ebiten.Image, image *ebiten.Image, gunImage *ebiten.Image, p *world.Player) {
	opt := &ebiten.DrawImageOptions{}
	drawCoords := p.Coords
	if renderData, ok := r.renderData[p.ID]; ok {
		// TODO: Adjust this based on velocity?
		correctionRate := float64(10)
		movement := math.Min(time.Since(r.lastDrawTime).Seconds()*correctionRate*60, correctionRate)
		if math.Abs(renderData.lastPlayerDrawCoords.X-drawCoords.X) > correctionRate {
			if drawCoords.X > renderData.lastPlayerDrawCoords.X {
				drawCoords.X = renderData.lastPlayerDrawCoords.X + movement
			} else {
				drawCoords.X = renderData.lastPlayerDrawCoords.X - movement
			}
		}

		if p.Jumping {
			correctionRate = 20
			movement = math.Min(time.Since(r.lastDrawTime).Seconds()*correctionRate*60, correctionRate)
		}
		if math.Abs(renderData.lastPlayerDrawCoords.Y-drawCoords.Y) > correctionRate {
			if drawCoords.Y > renderData.lastPlayerDrawCoords.Y {
				drawCoords.Y = renderData.lastPlayerDrawCoords.Y + movement
			} else {
				drawCoords.Y = renderData.lastPlayerDrawCoords.Y - movement
			}
		}
	} else {
		r.renderData[p.ID] = &RenderData{
			lastPlayerDrawCoords: drawCoords,
			prevAngle:            p.Angle,
		}
	}

	opt.GeoM.Translate(drawCoords.X, drawCoords.Y)
	opt.Filter = ebiten.FilterLinear
	_, playerHeight := image.Size()
	r.renderData[p.ID].lastPlayerDrawCoords = drawCoords
	screen.DrawImage(image, opt)
	RenderGun(screen, gunImage, drawCoords, p.Angle)
	debugString := fmt.Sprintf("%s\n(%0.0f,%0.0f)\n%d :: %d", p.ID, p.Coords.X, p.Coords.Y, p.Health, p.Ammo)
	ebitenutil.DebugPrintAt(screen, debugString, int(drawCoords.X), int(drawCoords.Y)+playerHeight)
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
