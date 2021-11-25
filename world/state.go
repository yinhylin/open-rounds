package world

import (
	"rounds/pb"
)

const NilTick int64 = -1

type State struct {
	Entities map[string]Entity
	Bullets  map[string]Bullet
	Tick     int64
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
