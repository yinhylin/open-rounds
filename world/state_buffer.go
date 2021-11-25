package world

import (
	"fmt"
	"log"
	"math"
	"rounds/pb"
)

type StateBuffer struct {
	states       []*State
	futureEvents map[int64][]*pb.ServerEvent
	index        int
	currentTick  int64
}

func (s *StateBuffer) ForEachBullet(callback func(string, *Bullet)) {
	if current := s.Current(); current != nil {
		for ID, bullet := range current.Bullets {
			callback(ID, &bullet)
		}
	}
}

func (s *StateBuffer) ForEachEntity(callback func(string, *Entity)) {
	if current := s.Current(); current != nil {
		for ID, entity := range current.Entities {
			callback(ID, &entity)
		}
	}
}

func (s *StateBuffer) CurrentTick() int64 {
	return s.currentTick
}

func newRingBuffer(maxCapacity int) []*State {
	states := make([]*State, maxCapacity)
	for i := range states {
		states[i] = NewState()
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
		futureEvents: make(map[int64][]*pb.ServerEvent),
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

	next := Simulate(current)
	s.Add(next)
	for _, event := range s.futureEvents[s.currentTick] {
		s.OnEvent(event)
	}
	delete(s.futureEvents, s.currentTick)
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
			s.states[index] = Simulate(currentState)
			currentState = s.states[index]
		})
		return nil
	}
	return fmt.Errorf("could not find tick. current=%d, server=%d", s.currentTick, tick)
}

func (s *StateBuffer) OnEvent(e *pb.ServerEvent) error {
	if e.Tick > s.currentTick {
		s.futureEvents[e.Tick] = append(s.futureEvents[e.Tick], e)
		return nil
	}

	switch e.Event.(type) {
	case *pb.ServerEvent_AddEntity:
		ID := e.GetAddEntity().Entity.Id
		return s.applyUpdate(e.Tick, func(state *State) {
			state.Entities[ID] = Entity{ID: ID}
		})

	case *pb.ServerEvent_RemoveEntity:
		ID := e.GetRemoveEntity().Id
		return s.applyUpdate(e.Tick, func(state *State) {
			delete(state.Entities, ID)
		})

	case *pb.ServerEvent_EntityEvents:
		msg := e.GetEntityEvents()
		return s.applyUpdate(e.Tick, func(state *State) {
			entity := state.Entities[msg.Id]
			entity.Intents = IntentsFromProto(msg.Intents)
			state.Entities[msg.Id] = entity
		})

	case *pb.ServerEvent_EntityAngle:
		msg := e.GetEntityAngle()
		return s.applyUpdate(e.Tick, func(state *State) {
			entity := state.Entities[msg.Id]
			entity.Angle = msg.Angle
			state.Entities[msg.Id] = entity
		})

	case *pb.ServerEvent_EntityShoot:
		msg := e.GetEntityShoot()
		// TODO: Validate can shoot etc. YOLO for now.
		return s.applyUpdate(e.Tick, func(state *State) {
			entity := state.Entities[msg.SourceId]
			state.Bullets[msg.Id] = Bullet{
				ID:     msg.Id,
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
	return fmt.Errorf("unhandled event=%+v\n", e)
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

	futureEvents := make(map[int64][]*pb.ServerEvent, len(p.FutureEvents))
	for _, event := range p.FutureEvents {
		futureEvents[event.Tick] = append(futureEvents[event.Tick], event)
	}
	return &StateBuffer{
		states:       states,
		futureEvents: futureEvents,
		index:        index,
		currentTick:  currentTick,
	}
}

func (s *StateBuffer) ToProto() *pb.StateBuffer {
	p := &pb.StateBuffer{
		MaxCapacity:  int64(cap(s.states)),
		States:       make([]*pb.State, 0, len(s.states)),
		FutureEvents: make([]*pb.ServerEvent, 0, len(s.futureEvents)),
	}
	for _, state := range s.states {
		p.States = append(p.States, state.ToProto())
	}
	for _, events := range s.futureEvents {
		for _, event := range events {
			p.FutureEvents = append(p.FutureEvents, event)
		}
	}
	return p
}
