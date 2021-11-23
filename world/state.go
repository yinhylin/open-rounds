package world

import (
	"fmt"
	"log"
	"rounds/pb"
)

const NilTick int64 = -1

type State struct {
	Entities map[string]Entity
	Tick     int64
}

func EntitiesFromProto(p []*pb.Entity) map[string]Entity {
	entities := make(map[string]Entity, len(p))
	for _, entity := range p {
		entities[entity.Id] = *EntityFromProto(entity)
	}
	return entities
}

func EntitiesToProto(e map[string]Entity) []*pb.Entity {
	p := make([]*pb.Entity, 0, len(e))
	for _, entity := range e {
		p = append(p, entity.ToProto())
	}
	return p
}

func StateFromProto(p *pb.State) *State {
	return &State{
		Tick:     p.Tick,
		Entities: EntitiesFromProto(p.EntityStates),
	}
}

func (s *State) ToProto() *pb.State {
	return &pb.State{
		Tick:         s.Tick,
		EntityStates: EntitiesToProto(s.Entities),
	}
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

type AngleUpdate struct {
	ID    string
	Angle float64
	Tick  int64
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

	for k := range a {
		if _, ok := b[k]; !ok {
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

type UpdateBuffer struct {
	Intents map[string]map[pb.Intents_Intent]struct{}
	Angles  map[string]float64
	Add     map[string]struct{}
	Remove  map[string]struct{}
}

func UpdateBufferFromProto(p *pb.UpdateBuffer) UpdateBuffer {
	u := NewUpdateBuffer()
	for _, intent := range p.Intents {
		u.Intents[intent.Id] = IntentsFromProto(intent.Intents)
	}
	for _, ID := range p.Add {
		u.Add[ID] = struct{}{}
	}
	for _, ID := range p.Remove {
		u.Remove[ID] = struct{}{}
	}
	return u
}

func (u *UpdateBuffer) ToProto(tick int64) *pb.UpdateBuffer {
	p := &pb.UpdateBuffer{
		Tick: tick,
	}
	for ID, intents := range u.Intents {
		p.Intents = append(p.Intents, &pb.EntityIntents{
			Id:      ID,
			Intents: IntentsToProto(intents),
		})
	}
	for ID := range u.Add {
		p.Add = append(p.Add, ID)
	}
	for ID := range u.Remove {
		p.Remove = append(p.Remove, ID)
	}
	return p
}

func NewUpdateBuffer() UpdateBuffer {
	return UpdateBuffer{
		Intents: make(map[string]map[pb.Intents_Intent]struct{}),
		Angles:  make(map[string]float64),
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

func StateBufferFromProto(p *pb.StateBuffer) *StateBuffer {
	states := newRingBuffer(int(p.MaxCapacity))
	currentTick := NilTick
	index := -1
	for i := range p.States {
		state := *StateFromProto(p.States[i])
		states[i] = state
		if state.Tick > currentTick {
			currentTick = state.Tick
			index = i
		}
	}

	updateBuffer := make(map[int64]UpdateBuffer, len(p.UpdateBuffers))
	for _, buffer := range p.UpdateBuffers {
		updateBuffer[buffer.Tick] = UpdateBufferFromProto(buffer)
	}
	return &StateBuffer{
		states:       states,
		updateBuffer: updateBuffer,
		index:        index,
		currentTick:  currentTick,
	}
}

func (s *StateBuffer) ToProto() *pb.StateBuffer {
	p := &pb.StateBuffer{
		MaxCapacity:   int64(cap(s.states)),
		States:        make([]*pb.State, 0, len(s.states)),
		UpdateBuffers: make([]*pb.UpdateBuffer, 0, len(s.updateBuffer)),
	}
	for _, state := range s.states {
		p.States = append(p.States, state.ToProto())
	}
	for tick, buffer := range s.updateBuffer {
		p.UpdateBuffers = append(p.UpdateBuffers, buffer.ToProto(tick))
	}
	return p
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

func newRingBuffer(maxCapacity int) []State {
	states := make([]State, maxCapacity)
	for i := range states {
		states[i].Tick = NilTick
	}
	return states
}

func (s *StateBuffer) Clear() {
	s.states = newRingBuffer(cap(s.states))
	s.currentTick = NilTick
}

func NewStateBuffer(maxCapacity int) *StateBuffer {
	return &StateBuffer{
		states:       newRingBuffer(maxCapacity),
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

	next := Simulate(*current, s.updateBuffer[s.currentTick+1])
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
		nextIndex := (index + i) % cap(s.states)
		if s.states[index].Tick > s.states[nextIndex].Tick {
			log.Fatal(s.states[i].Tick, s.states[nextIndex].Tick)
		}
		callback((index + i) % cap(s.states))
	}
}

func (s *StateBuffer) applyUpdate(tick int64, callback func(State) State) error {
	if s.currentTick < tick {
		log.Fatal(tick, s.currentTick)
	}

	for i, state := range s.states {
		if state.Tick != tick {
			continue
		}
		s.states[i] = callback(s.states[i])
		currentState := &s.states[i]
		// Re-simulate.
		s.walkNextStates(i, int(s.currentTick-state.Tick), func(index int) {
			s.states[index] = Simulate(*currentState, s.updateBuffer[currentState.Tick+1])
			currentState = &s.states[index]
		})
		return nil
	}
	return fmt.Errorf("could not find tick. current=%d, server=%d", s.currentTick, tick)
}

func (s *StateBuffer) modifyUpdateBuffer(tick int64, callback func(UpdateBuffer) UpdateBuffer) {
	if _, ok := s.updateBuffer[tick]; !ok {
		s.updateBuffer[tick] = NewUpdateBuffer()
	}
	s.updateBuffer[tick] = callback(s.updateBuffer[tick])
}

func (s *StateBuffer) AddEntity(msg *AddEntity) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Add[msg.ID] = struct{}{}
			return buffer
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(existing State) State {
		existing.Entities[msg.ID] = Entity{ID: msg.ID}
		fmt.Printf("%+v\n", existing.Entities)
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}

func (s *StateBuffer) RemoveEntity(msg *RemoveEntity) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Remove[msg.ID] = struct{}{}
			return buffer
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(existing State) State {
		delete(existing.Entities, msg.ID)
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}

func (s *StateBuffer) ApplyIntents(msg *IntentsUpdate) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Intents[msg.ID] = msg.Intents
			return buffer
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(existing State) State {
		entity := existing.Entities[msg.ID]
		entity.Intents = msg.Intents
		existing.Entities[msg.ID] = entity
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}

func (s *StateBuffer) ApplyAngle(msg *AngleUpdate) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer UpdateBuffer) UpdateBuffer {
			buffer.Angles[msg.ID] = msg.Angle
			return buffer
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(existing State) State {
		entity := existing.Entities[msg.ID]
		entity.Angle = msg.Angle
		existing.Entities[msg.ID] = entity
		return State{
			Entities: existing.Entities,
			Tick:     msg.Tick,
		}
	})
}
