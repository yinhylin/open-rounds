package world

import (
	"log"
	"rounds/pb"
)

func updateEntity(e *Entity) {
	const speed = 4
	// TODO: Don't overwrite velocity.
	var velocity Vector
	for action := range e.Intents {
		switch action {
		case pb.Intents_MOVE_UP:
			velocity.Y -= speed
		case pb.Intents_MOVE_DOWN:
			velocity.Y += speed
		case pb.Intents_MOVE_LEFT:
			velocity.X -= speed
		case pb.Intents_MOVE_RIGHT:
			velocity.X += speed
		}
	}

	// TODO: This needs to move elsewhere and have collision checking.
	e.Velocity = velocity
	e.Coords.X += e.Velocity.X
	e.Coords.Y += e.Velocity.Y
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
