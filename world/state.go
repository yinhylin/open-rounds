package world

import (
	"log"
	"rounds/pb"
)

const NilTick = -1

type State struct {
	Entities map[string]Entity
	Tick     int64
}

type IntentsUpdate struct {
	ID      string
	Intents map[pb.Intents_Intent]struct{}
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

func IntentsEqual(a, b map[pb.Intents_Intent]struct{}) bool {
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
	return a.Coords == b.Coords && a.Velocity == b.Velocity && IntentsEqual(a.Intents, b.Intents)
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

func (s *State) Next(updateBuffer UpdateBuffer) State {
	next := *s
	next.Entities = make(map[string]Entity, len(s.Entities))

	// Add
	for ID := range updateBuffer.Add {
		next.Entities[ID] = Entity{ID: ID}
	}

	// Update
	for ID, entity := range s.Entities {
		if _, ok := updateBuffer.Remove[ID]; ok {
			// Remove
			continue
		}
		if intents, ok := updateBuffer.Intents[ID]; ok {
			entity.Intents = intents
		}
		entity.Update()
		next.Entities[ID] = entity
	}

	next.Tick++
	return next
}

type UpdateBuffer struct {
	Intents map[string]map[pb.Intents_Intent]struct{}
	Add     map[string]struct{}
	Remove  map[string]struct{}
}

var emptyUpdateBuffer UpdateBuffer

func NewUpdateBuffer() UpdateBuffer {
	return UpdateBuffer{
		Intents: make(map[string]map[pb.Intents_Intent]struct{}),
		Add:     make(map[string]struct{}),
		Remove:  make(map[string]struct{}),
	}
}

type StateBuffer struct {
	states       []State
	updateBuffer map[int64]UpdateBuffer
	index        int
	currentTick  int64
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
		states[i].Tick = NilTick
	}
	return &StateBuffer{
		states:       states,
		updateBuffer: make(map[int64]UpdateBuffer),
		currentTick:  NilTick,
	}
}

func (s *StateBuffer) Add(state *State) {
	index := (s.index + 1) % cap(s.states)
	if s.states[s.index].Tick == NilTick {
		index = s.index
	}
	s.index = index
	s.states[index] = *state
	s.currentTick = state.Tick
}

func (s *StateBuffer) Next() *State {
	current := s.Current()
	if current == nil {
		return nil
	}

	next := current.Next(s.updateBuffer[s.currentTick+1])
	s.Add(&next)
	delete(s.updateBuffer, s.currentTick)
	return &next
}

func (s *StateBuffer) Current() *State {
	current := s.states[s.index]
	if current.Tick == NilTick {
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
		log.Println("fast forward ", s.currentTick, tick)
	}

	if s.currentTick < tick {
		// need to buffer this xD
		// log.Fatal(tick, s.currentTick)
		return
	}

	for i, state := range s.states {
		if state.Tick != tick {
			continue
		}
		s.states[i] = callback(s.states[i])
		currentState := &s.states[i]

		// Re-simulate.
		s.walkNextStates(i, int(s.currentTick-state.Tick), func(index int) {
			s.states[index] = currentState.Next(s.updateBuffer[currentState.Tick+1])
			currentState = &s.states[index]
		})
		return
	}
}

func (s *StateBuffer) modifyUpdateBuffer(tick int64, callback func(UpdateBuffer) UpdateBuffer) {
	if _, ok := s.updateBuffer[tick]; !ok {
		s.updateBuffer[tick] = NewUpdateBuffer()
	}
	s.updateBuffer[tick] = callback(s.updateBuffer[tick])
}

func (s *StateBuffer) AddEntity(msg *AddEntity) {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Add[msg.ID] = struct{}{}
			return buffer
		})
		return
	}
	s.applyUpdate(msg.Tick, func(existing State) State {
		existing.Entities[msg.ID] = Entity{ID: msg.ID}
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}

func (s *StateBuffer) RemoveEntity(msg *RemoveEntity) {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Remove[msg.ID] = struct{}{}
			return buffer
		})
		return
	}
	s.applyUpdate(msg.Tick, func(existing State) State {
		delete(existing.Entities, msg.ID)
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}

func (s *StateBuffer) ApplyIntents(msg *IntentsUpdate) {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Intents[msg.ID] = msg.Intents
			return buffer
		})
		return
	}

	s.applyUpdate(msg.Tick, func(existing State) State {
		entity := existing.Entities[msg.ID]
		entity.Intents = msg.Intents
		existing.Entities[msg.ID] = entity
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}
