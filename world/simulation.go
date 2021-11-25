package world

import (
	"math"
	"rounds/pb"
)

func updateEntity(e *Entity) {
	const speed = 10
	// TODO: Don't overwrite velocity.
	jump := false
	var velocity Vector
	for action := range e.Intents {
		switch action {
		case pb.Intents_JUMP:
			jump = true
		case pb.Intents_MOVE_LEFT:
			velocity.X -= speed
		case pb.Intents_MOVE_RIGHT:
			velocity.X += speed
		}
	}

	// gravity
	velocity.Y = math.Min(e.Velocity.Y+2, 16)
	e.Velocity = velocity
	e.Coords.X += e.Velocity.X
	e.Coords.Y += e.Velocity.Y

	// TODO: This needs to be proper collision detection but yolo prototyping.
	// TODO: Finish Map.
	if e.Coords.Y > 500 {
		e.Coords.Y = 500
		if !jump {
			// lol bounce
			e.Velocity.Y = -math.Abs(e.Velocity.Y / 1.25)
		} else {
			e.Velocity.Y = 0
		}
	}

	if e.Coords.Y == 500 && jump {
		e.Velocity.Y -= 32
	}
}

func updateBullet(b *Bullet) bool {
	b.Velocity.Y = math.Min(b.Velocity.Y+1, 32)
	b.Coords.X += b.Velocity.X
	b.Coords.Y += b.Velocity.Y
	return b.Coords.Y < 720
}

func Simulate(s *State) *State {
	next := &State{
		Entities: make(map[string]Entity, len(s.Entities)),
		Bullets:  make(map[string]Bullet, len(s.Bullets)),
		Tick:     s.Tick + 1,
	}

	// Update
	for ID, bullet := range s.Bullets {
		if updateBullet(&bullet) {
			next.Bullets[ID] = bullet
		}
	}
	for ID, entity := range s.Entities {
		updateEntity(&entity)
		next.Entities[ID] = entity
	}
	return next
}
