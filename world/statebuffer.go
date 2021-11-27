package world

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/sailormoon/open-rounds/pb"
)

type StateBuffer struct {
	states       []*State
	futureEvents map[int64][]*pb.ServerEvent
	index        int
	currentTick  int64
	m            *Map
}

func NewStateBuffer(maxCapacity int, m *Map) *StateBuffer {
	return &StateBuffer{
		states:       newRingBuffer(maxCapacity),
		futureEvents: make(map[int64][]*pb.ServerEvent),
		currentTick:  NilTick,
		m:            m,
	}
}

func (s *StateBuffer) OnEvent(e *pb.ServerEvent) error {
	if e.GetPlayer() == nil {
		return errors.New("no player details")
	}
	details := e.Player

	if details.Tick > s.currentTick {
		s.futureEvents[details.Tick] = append(s.futureEvents[details.Tick], e)
		return nil
	}

	switch e.Event.(type) {
	case *pb.ServerEvent_AddPlayer:
		return s.applyUpdate(details.Tick, func(state *State) {
			state.Players[details.Id] = Player{ID: details.Id}
		})

	case *pb.ServerEvent_RemovePlayer:
		return s.applyUpdate(details.Tick, func(state *State) {
			delete(state.Players, details.Id)
		})

	case *pb.ServerEvent_Intents:
		msg := e.GetIntents()
		return s.applyUpdate(details.Tick, func(state *State) {
			player := state.Players[details.Id]
			player.Intents = IntentsFromProtoSlice(msg.Intents)
			state.Players[details.Id] = player
		})

	case *pb.ServerEvent_Angle:
		msg := e.GetAngle()
		return s.applyUpdate(details.Tick, func(state *State) {
			player := state.Players[details.Id]
			player.Angle = msg.Angle
			state.Players[details.Id] = player
		})
	}
	return fmt.Errorf("unhandled event=%+v\n", e)
}

func (s *StateBuffer) Next() *State {
	current := s.Current()
	if current == nil {
		return nil
	}

	next := Simulate(current, s.m)
	s.Add(next)
	for _, event := range s.futureEvents[s.currentTick] {
		s.OnEvent(event)
	}
	delete(s.futureEvents, s.currentTick)
	return next
}

func (s *StateBuffer) Current() *State {
	if current := s.states[s.index]; current.Tick != NilTick {
		return current
	}
	return nil
}

func (s *StateBuffer) Map() *Map {
	return s.m
}

func (s *StateBuffer) Clear() {
	s.states = newRingBuffer(cap(s.states))
	s.currentTick = NilTick
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

func (s *StateBuffer) ForEachBullet(callback func(*Bullet)) {
	if current := s.Current(); current != nil {
		for _, bullet := range current.Bullets {
			callback(&bullet)
		}
	}
}

func (s *StateBuffer) ForEachPlayer(callback func(string, *Player)) {
	if current := s.Current(); current != nil {
		sortedIDs := make([]string, 0, len(current.Players))
		for ID := range current.Players {
			sortedIDs = append(sortedIDs, ID)
		}
		sort.Slice(sortedIDs, func(i, j int) bool {
			l := current.Players[sortedIDs[i]]
			r := current.Players[sortedIDs[j]]
			if l.Coords.Y == r.Coords.Y {
				if l.Coords.X == r.Coords.X {
					return l.ID > r.ID
				}
				return l.Coords.X > r.Coords.X
			}
			return l.Coords.Y > r.Coords.Y
		})
		for _, ID := range sortedIDs {
			player := current.Players[ID]
			callback(ID, &player)
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
			s.states[index] = Simulate(currentState, s.m)
			currentState = s.states[index]
		})
		return nil
	}
	return fmt.Errorf("could not find tick. current=%d, server=%d", s.currentTick, tick)
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
		m:            MapFromProto(p.Map),
	}
}

func (s *StateBuffer) ToProto() *pb.StateBuffer {
	p := &pb.StateBuffer{
		MaxCapacity:  int64(cap(s.states)),
		States:       make([]*pb.State, 0, len(s.states)),
		FutureEvents: make([]*pb.ServerEvent, 0, len(s.futureEvents)),
		Map:          s.m.ToProto(),
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
