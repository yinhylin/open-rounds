package world

import (
	"math"

	"github.com/sailormoon/open-rounds/pb"
)

func updatePlayer(e *Player, s *State, m *Map) {
	const speed = 10
	// TODO: Don't overwrite velocity.
	jump := false
	shoot := false
	var velocity Vector
	for action := range e.Intents {
		switch action {
		case pb.Intents_JUMP:
			jump = true
		case pb.Intents_SHOOT:
			shoot = true
		case pb.Intents_MOVE_LEFT:
			velocity.X -= speed
		case pb.Intents_MOVE_RIGHT:
			velocity.X += speed
		}
	}

	// gravity
	velocity.Y = math.Min(e.Velocity.Y+2, 16)
	e.Velocity = velocity

	grounded := false
	const size = tileSize - 1
	// horizontal scan
	if e.Velocity.X != 0 {
		for _, dy := range []float64{0, size} {
			toCheck := Vector{
				X: e.Coords.X + e.Velocity.X,
				Y: e.Coords.Y + dy,
			}
			if e.Velocity.X > 0 {
				toCheck.X += size
			}
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			if err != nil {
				continue
			}
			if tile.Dense {
				if e.Velocity.X > 0 {
					e.Coords.X = float64((x - 1) * 32)
				} else if e.Velocity.X < 0 {
					e.Coords.X = float64((x + 1) * 32)
				}
				e.Velocity.X = 0
				break
			}
		}
	}

	// vertical scan
	if e.Velocity.Y != 0 {
		for _, dx := range []float64{0, size} {
			toCheck := Vector{
				X: e.Coords.X + dx,
				Y: e.Coords.Y + e.Velocity.Y,
			}
			if e.Velocity.Y > 0 {
				toCheck.Y += size
			}
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			if err != nil {
				continue
			}
			if tile.Dense {
				if e.Velocity.Y > 0 {
					e.Coords.Y = float64((y - 1) * 32)
					grounded = true
				} else if e.Velocity.Y < 0 {
					e.Coords.Y = float64((y + 1) * 32)
				}
				e.Velocity.Y = 0
			}
		}
	}

	e.Coords.X += e.Velocity.X
	e.Coords.Y += e.Velocity.Y

	if grounded && jump {
		e.Velocity.Y -= 32
	}

	if shoot {
		s.Bullets = append(s.Bullets, Bullet{
			Coords: e.Coords,
			Velocity: Vector{
				// TODO: Use gun constants and stuff.
				X: -math.Cos(e.Angle) * 30,
				Y: -math.Sin(e.Angle) * 30,
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
