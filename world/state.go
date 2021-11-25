package world

import (
	"rounds/pb"
)

const NilTick int64 = -1

type State struct {
	Players map[string]Player
	Bullets []Bullet
	Tick    int64
}

func NewState() *State {
	return &State{
		Tick:    NilTick,
		Players: make(map[string]Player),
	}
}

func StateFromProto(p *pb.State) *State {
	return &State{
		Tick:    p.Tick,
		Players: PlayersFromProto(p.PlayerStates),
	}
}

func (s *State) ToProto() *pb.State {
	return &pb.State{
		Tick:         s.Tick,
		PlayerStates: PlayersToProto(s.Players),
	}
}
