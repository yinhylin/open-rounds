package world

import (
	"log"
	"math"
	"rounds/pb"
)

func updateEntity(e *Entity) {
	const speed = 8
	// TODO: Don't overwrite velocity.
	jump := false
	var velocity Vector
	for action := range e.Intents {
		switch action {
		case pb.Intents_MOVE_UP:
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

func Simulate(s State, u UpdateBuffer) State {
	next := s
	next.Entities = make(map[string]Entity, len(s.Entities))
	// Add
	for ID := range u.Add {
		log.Println("add ", ID)
		next.Entities[ID] = Entity{ID: ID}
	}

	// Update
	for ID, entity := range s.Entities {
		if _, ok := u.Remove[ID]; ok {
			log.Println("remove ", ID)
			// Remove
			continue
		}
		updateEntity(&entity)
		if intents, ok := u.Intents[ID]; ok {
			entity.Intents = intents
		}
		next.Entities[ID] = entity
	}

	next.Tick++
	return next
}
