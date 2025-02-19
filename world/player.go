package world

import "github.com/sailormoon/open-rounds/pb"

// TODO: This all needs a giant refactor to dedupe things. :(
type Player struct {
	ID       string
	Coords   Vector
	Velocity Vector
	Intents  map[pb.Intents_Intent]struct{}
	Angle    float64
}

func IntentsToProto(actions map[pb.Intents_Intent]struct{}) *pb.Intents {
	var events []pb.Intents_Intent
	for action := range actions {
		events = append(events, action)
	}
	return &pb.Intents{
		Intents: events,
	}
}

func IntentsFromProtoSlice(intents []pb.Intents_Intent) map[pb.Intents_Intent]struct{} {
	actions := make(map[pb.Intents_Intent]struct{})
	for _, action := range intents {
		actions[action] = struct{}{}
	}
	return actions
}

func IntentsToProtoSlice(actions map[pb.Intents_Intent]struct{}) []pb.Intents_Intent {
	var events []pb.Intents_Intent
	for action := range actions {
		events = append(events, action)
	}
	return events
}

func IntentsFromProto(a *pb.Intents) map[pb.Intents_Intent]struct{} {
	actions := make(map[pb.Intents_Intent]struct{})
	for _, action := range a.Intents {
		actions[action] = struct{}{}
	}
	return actions
}

func (e *Player) ToProto() *pb.Player {
	return &pb.Player{
		Id:       e.ID,
		Position: e.Coords.ToProto(),
		Velocity: e.Velocity.ToProto(),
		Intents:  IntentsToProto(e.Intents),
		Angle:    e.Angle,
	}
}

func PlayerFromProto(e *pb.Player) *Player {
	if e == nil {
		return nil
	}
	return &Player{
		ID:       e.Id,
		Coords:   VectorFromProto(e.Position),
		Velocity: VectorFromProto(e.Velocity),
		Intents:  IntentsFromProto(e.Intents),
		Angle:    e.Angle,
	}
}

func PlayersFromProto(p []*pb.Player) map[string]Player {
	players := make(map[string]Player, len(p))
	for _, player := range p {
		players[player.Id] = *PlayerFromProto(player)
	}
	return players
}

func PlayersToProto(e map[string]Player) []*pb.Player {
	p := make([]*pb.Player, 0, len(e))
	for _, player := range e {
		p = append(p, player.ToProto())
	}
	return p
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
