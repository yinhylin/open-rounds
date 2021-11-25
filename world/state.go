package world

import (
	"fmt"
	"log"
	"math"
	"rounds/pb"
)

const NilTick int64 = -1

type State struct {
	Entities map[string]Entity
	Bullets  map[string]Bullet
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

type AddBullet struct {
	Source string
	ID     string
	Tick   int64
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

type StateBuffer struct {
	states       []*State
	updateBuffer map[int64]*UpdateBuffer
	index        int
	currentTick  int64
}

func StateBufferFromProto(p *pb.StateBuffer) *StateBuffer {
	states := newRingBuffer(int(p.MaxCapacity))
	currentTick := NilTick
	index := -1
	for i := range p.States {
		state := StateFromProto(p.States[i])
		states[i] = state
		if state.Tick > currentTick {
			currentTick = state.Tick
			index = i
		}
	}

	updateBuffer := make(map[int64]*UpdateBuffer, len(p.UpdateBuffers))
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

func (s *StateBuffer) ForEachBullet(callback func(string, *Bullet)) {
	current := s.Current()
	if current == nil {
		return
	}

	for ID, bullet := range current.Bullets {
		callback(ID, &bullet)
	}
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

func newRingBuffer(maxCapacity int) []*State {
	states := make([]*State, maxCapacity)
	for i := range states {
		states[i] = &State{
			Tick: NilTick,
		}
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
		updateBuffer: make(map[int64]*UpdateBuffer),
		currentTick:  NilTick,
	}
}

func (s *StateBuffer) Add(state *State) {
	index := (s.index + 1) % cap(s.states)
	if s.states[s.index].Tick == NilTick {
		index = s.index
	}
	s.index = index
	s.states[index] = state
	s.currentTick = state.Tick
}

func (s *StateBuffer) Next() *State {
	current := s.Current()
	if current == nil {
		return nil
	}

	next := Simulate(current, s.updateBuffer[s.currentTick+1])
	s.Add(next)
	delete(s.updateBuffer, s.currentTick)
	return next
}

func (s *StateBuffer) Current() *State {
	current := s.states[s.index]
	if current.Tick == NilTick {
		return nil
	}
	return current
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

func (s *StateBuffer) applyUpdate(tick int64, callback func(*State)) error {
	if s.currentTick < tick {
		log.Fatal(tick, s.currentTick)
	}

	for i, state := range s.states {
		if state.Tick != tick {
			continue
		}
		callback(s.states[i])
		currentState := s.states[i]
		// Re-simulate.
		s.walkNextStates(i, int(s.currentTick-state.Tick), func(index int) {
			s.states[index] = Simulate(currentState, s.updateBuffer[currentState.Tick+1])
			currentState = s.states[index]
		})
		return nil
	}
	return fmt.Errorf("could not find tick. current=%d, server=%d", s.currentTick, tick)
}

func (s *StateBuffer) modifyUpdateBuffer(tick int64, callback func(*UpdateBuffer)) {
	if _, ok := s.updateBuffer[tick]; !ok {
		s.updateBuffer[tick] = NewUpdateBuffer()
	}
	callback(s.updateBuffer[tick])
}

func (s *StateBuffer) AddEntity(msg *AddEntity) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer *UpdateBuffer) {
			buffer.Add[msg.ID] = struct{}{}
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(state *State) {
		state.Entities[msg.ID] = Entity{ID: msg.ID}
	})
}

func (s *StateBuffer) RemoveEntity(msg *RemoveEntity) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer *UpdateBuffer) {
			buffer.Remove[msg.ID] = struct{}{}
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(state *State) {
		delete(state.Entities, msg.ID)
	})
}

func (s *StateBuffer) ApplyIntents(msg *IntentsUpdate) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer *UpdateBuffer) {
			buffer.Intents[msg.ID] = msg.Intents
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(state *State) {
		entity := state.Entities[msg.ID]
		entity.Intents = msg.Intents
		state.Entities[msg.ID] = entity
	})
}

func (s *StateBuffer) ApplyAngle(msg *AngleUpdate) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer *UpdateBuffer) {
			buffer.Angles[msg.ID] = msg.Angle
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(state *State) {
		entity := state.Entities[msg.ID]
		entity.Angle = msg.Angle
		state.Entities[msg.ID] = entity
	})
}

func (s *StateBuffer) AddBullet(msg *AddBullet) error {
	if msg.Tick > s.currentTick {
		s.modifyUpdateBuffer(msg.Tick, func(buffer *UpdateBuffer) {
			buffer.Shots[msg.Source] = append(buffer.Shots[msg.Source], msg.ID)
		})
		return nil
	}

	return s.applyUpdate(msg.Tick, func(state *State) {
		// TODO: Validate can shoot etc. YOLO for now.
		entity := state.Entities[msg.Source]
		state.Bullets[msg.ID] = Bullet{
			ID:     msg.ID,
			Coords: entity.Coords,
			Velocity: Vector{
				// TODO: Use gun constants and stuff.
				X: -math.Cos(entity.Angle)*30 + entity.Velocity.X,
				Y: -math.Sin(entity.Angle)*30 + entity.Velocity.Y,
			},
			Angle: entity.Angle,
		}
	})
}
