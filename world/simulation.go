package world

import (
	"math"

	"github.com/sailormoon/open-rounds/pb"
)

func updatePlayer(p *Player, s *State, m *Map) {
	const speed = 10
	// TODO: Don't overwrite velocity.
	shoot := false
	jump := false
	var velocity Vector
	for action := range p.Intents {
		switch action {
		case pb.Intents_JUMP:
			jump = true
			if p.JumpReleasedTick >= p.JumpTick {
				p.JumpTick = s.Tick
			}
		case pb.Intents_SHOOT:
			shoot = true
		case pb.Intents_MOVE_LEFT:
			velocity.X -= speed
		case pb.Intents_MOVE_RIGHT:
			velocity.X += speed
		}
	}

	if !jump && p.JumpReleasedTick < p.JumpTick {
		p.JumpReleasedTick = s.Tick
	}

	velocity.Y = math.Min(p.Velocity.Y+2, 16)
	p.Velocity = velocity

	const size = tileSize - 1
	// horizontal scan
	if p.Velocity.X != 0 {
		for _, dy := range []float64{0, size} {
			toCheck := Vector{
				X: p.Coords.X + p.Velocity.X,
				Y: p.Coords.Y + dy,
			}
			if p.Velocity.X > 0 {
				toCheck.X += size
			}
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			dense := err != nil || tile.Dense
			if dense {
				if p.Velocity.X > 0 {
					p.Coords.X = float64((x - 1) * 32)
				} else if p.Velocity.X < 0 {
					p.Coords.X = float64((x + 1) * 32)
				}
				p.Velocity.X = 0
				break
			}
		}
	}

	// vertical scan
	if p.Velocity.Y != 0 {
		for _, dx := range []float64{0, size} {
			toCheck := Vector{
				X: p.Coords.X + dx,
				Y: p.Coords.Y + p.Velocity.Y,
			}
			if p.Velocity.Y > 0 {
				toCheck.Y += size
			}
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			dense := err != nil || tile.Dense
			if dense {
				if p.Velocity.Y > 0 {
					p.Coords.Y = float64((y - 1) * 32)
					p.GroundedTick = s.Tick
					p.Jumping = false
				} else if p.Velocity.Y < 0 {
					p.Coords.Y = float64((y + 1) * 32)
				}
				p.Velocity.Y = 0
			}
		}
	}

	p.Coords.X += p.Velocity.X
	p.Coords.Y += p.Velocity.Y

	wantJump := s.Tick-p.JumpTick < 5
	if wantJump && math.Abs(float64(p.GroundedTick-s.Tick)) < 5 && !p.Jumping {
		p.Velocity.Y = -32
		p.Jumping = true
	}

	if p.JumpReleasedTick > p.JumpTick && p.Jumping && p.Velocity.Y < -8 {
		p.Velocity.Y = -8
	}

	if shoot {
		s.Bullets = append(s.Bullets, Bullet{
			Coords: p.Coords,
			Velocity: Vector{
				// TODO: Use gun constants and stuff.
				X: -math.Cos(p.Angle) * 30,
				Y: -math.Sin(p.Angle) * 30,
			},
		})
	}
}

func updateBullet(b *Bullet) bool {
	b.Velocity.Y = math.Min(b.Velocity.Y+1, 32)
	b.Coords.X += b.Velocity.X
	b.Coords.Y += b.Velocity.Y
	return b.Coords.Y < 720
}

func Simulate(s *State, m *Map) *State {
	next := &State{
		Players: make(map[string]Player, len(s.Players)),
		Bullets: make([]Bullet, 0, len(s.Bullets)),
		Tick:    s.Tick + 1,
	}
	for _, bullet := range s.Bullets {
		if updateBullet(&bullet) {
			next.Bullets = append(next.Bullets, bullet)
		}
	}
	for ID, player := range s.Players {
		updatePlayer(&player, next, m)
		next.Players[ID] = player
	}
	return next
}
