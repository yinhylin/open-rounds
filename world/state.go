package world

import (
	"github.com/sailormoon/open-rounds/pb"
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
		Bullets: BulletsFromProto(p.Bullets),
	}
}

func (s *State) ToProto() *pb.State {
	return &pb.State{
		Tick:         s.Tick,
		PlayerStates: PlayersToProto(s.Players),
		Bullets:      BulletsToProto(s.Bullets),
	}
}
