package world

import (
	"log"
	"rounds/pb"
)

const nilTick = -1

type State struct {
	Simulated bool
	Entities  map[string]Entity
	Tick      int64
}

type ActionsUpdate struct {
	ID      string
	Actions map[pb.Actions_Event]struct{}
	Tick    int64
}

type EntityUpdate struct {
	ID     string
	Entity Entity
	Tick   int64
}

type AddEntity struct {
	ID   string
	Tick int64
}

type RemoveEntity struct {
	ID   string
	Tick int64
}

func actionsEqual(a, b map[pb.Actions_Event]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func entityEqual(a, b Entity) bool {
	return a.Coords == b.Coords && a.Velocity == b.Velocity && actionsEqual(a.Actions, b.Actions)
}

func entitiesEqual(a, b map[string]Entity) bool {
	if len(a) != len(b) {
		return false
	}
	for ID, entity := range a {
		if !entityEqual(entity, b[ID]) {
			return false
		}
	}
	return true
}

func (s *State) Next() State {
	next := *s
	next.Entities = make(map[string]Entity, len(s.Entities))
	for ID, entity := range s.Entities {
		entity.Update()
		next.Entities[ID] = entity
	}
	next.Tick++
	next.Simulated = true
	return next
}

type StateBuffer struct {
	states                 []State
	index                  int
	currentServerTick      int64
	currentServerTickIndex int
	currentTick            int64
}

func (s *StateBuffer) ForEachEntity(callback func(string, *Entity)) {
	current := s.Current()
	if current == nil {
		return
	}

	for ID, entity := range current.Entities {
		callback(ID, &entity)
	}
}

func (s *StateBuffer) CurrentTick() int64 {
	return s.currentTick
}

func NewStateBuffer(maxCapacity int) *StateBuffer {
	states := make([]State, maxCapacity, maxCapacity)
	for i := range states {
		states[i].Tick = nilTick
	}
	return &StateBuffer{
		states:            states,
		currentServerTick: nilTick,
		currentTick:       nilTick,
	}
}

func (s *StateBuffer) Add(state *State) {
	index := (s.index + 1) % cap(s.states)
	if s.states[s.index].Tick == nilTick {
		index = s.index
	}
	s.index = index
	s.states[index] = *state
	if state.Simulated {
		s.currentTick = state.Tick
	} else {
		s.currentServerTick = state.Tick
		s.currentServerTickIndex = index
	}
}

func (s *StateBuffer) Next() *State {
	if int(s.currentTick-s.currentServerTick+1) >= cap(s.states) {
		// Simulated too far.
		return nil
	}

	current := s.Current()
	if current == nil {
		return nil
	}

	next := current.Next()
	s.Add(&next)
	return &next
}

func (s *StateBuffer) Current() *State {
	current := s.states[s.index]
	if current.Tick == nilTick {
		return nil
	}
	return &current
}

func (s *StateBuffer) walkNextStates(index int, steps int, callback func(int)) {
	for i := 1; i <= steps; i++ {
		callback((index + i) % cap(s.states))
	}
}

func (s *StateBuffer) applyUpdate(tick int64, callback func(State) State) {
	// Fast forward until we have a state.
	for s.currentTick < tick && s.Next() != nil {
	}

	if s.currentTick < tick {
		log.Fatal(tick, s.currentTick)
	}

	for i, state := range s.states {
		if state.Tick != tick {
			continue
		}
		s.currentServerTick = tick
		s.currentServerTickIndex = i

		s.states[i] = callback(s.states[i])
		currentState := &s.states[i]

		// Re-simulate.
		s.walkNextStates(i, int(s.currentTick-s.currentServerTick), func(index int) {
			s.states[index] = currentState.Next()
			currentState = &s.states[index]
		})
		return
	}
}

func (s *StateBuffer) AddEntity(msg *AddEntity) {
	s.applyUpdate(msg.Tick, func(existing State) State {
		existing.Entities[msg.ID] = Entity{ID: msg.ID}
		return State{
			Simulated: false,
			Entities:  existing.Entities,
			Tick:      msg.Tick,
		}
	})
}

func (s *StateBuffer) RemoveEntity(msg *RemoveEntity) {
	s.applyUpdate(msg.Tick, func(existing State) State {
		delete(existing.Entities, msg.ID)
		return State{
			Simulated: false,
			Entities:  existing.Entities,
			Tick:      msg.Tick,
		}
	})
}

func (s *StateBuffer) ApplyActions(msg *ActionsUpdate) {
	s.applyUpdate(msg.Tick, func(existing State) State {
		entity := existing.Entities[msg.ID]
		entity.Actions = msg.Actions
		existing.Entities[msg.ID] = entity
		return State{
			Simulated: false,
			Entities:  existing.Entities,
			Tick:      msg.Tick,
		}
	})
}

func (s *StateBuffer) ApplySimulatedActions(msg *ActionsUpdate) {
	s.applyUpdate(msg.Tick, func(existing State) State {
		entity := existing.Entities[msg.ID]
		entity.Actions = msg.Actions
		existing.Entities[msg.ID] = entity
		return State{
			Simulated: existing.Simulated,
			Entities:  existing.Entities,
			Tick:      msg.Tick,
		}
	})
}
