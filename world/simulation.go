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
	// horizontal scan
	{
		toCheck := e.Coords
		toCheck.X += e.Velocity.X
		if e.Velocity.X > 0 {
			toCheck.X += tileSize
		}
		for _, dy := range []float64{0, tileSize} {
			toCheck.Y += dy
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			if err != nil {
				continue
			}
			if tile.Dense {
				if e.Velocity.X > 0 {
					e.Coords.X += float64(32-int64(e.Coords.X)%32) - 1
				} else if e.Velocity.X < 0 {
					e.Coords.X -= float64(int64(e.Coords.X)%32) - 1
				}
				e.Velocity.X = 0
			}
		}
	}

	// vertical scan
	{
		toCheck := e.Coords
		toCheck.Y += e.Velocity.Y
		if e.Velocity.Y > 0 {
			toCheck.Y += tileSize
		}
		for _, dx := range []float64{0, tileSize} {
			toCheck.X += dx
			x, y := toCheck.ToTileCoordinates()
			tile, err := m.At(x, y)
			if err != nil {
				continue
			}
			if tile.Dense {
				if e.Velocity.Y > 0 {
					grounded = true
					e.Coords.Y += float64(32-int64(e.Coords.Y)%32) - 1
				} else if e.Velocity.Y < 0 {
					e.Coords.Y -= float64(int64(e.Coords.Y)%32) - 1
				}
				e.Velocity.Y = 0
			}
		}
	}
	e.Coords.X += e.Velocity.X
	e.Coords.Y += e.Velocity.Y

	// TODO: This needs to be proper collision detection but yolo prototyping.
	// TODO: Finish Map.
	if grounded {
		e.Velocity.Y = 0
	}

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
