package world

import "log"

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
		entity.Update()
		if intents, ok := u.Intents[ID]; ok {
			entity.Intents = intents
		}
		next.Entities[ID] = entity
	}

	next.Tick++
	return next
}
