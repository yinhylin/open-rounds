package world

import "rounds/pb"

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
	entities := make(map[string]Player, len(p))
	for _, entity := range p {
		entities[entity.Id] = *PlayerFromProto(entity)
	}
	return entities
}

func PlayersToProto(e map[string]Player) []*pb.Player {
	p := make([]*pb.Player, 0, len(e))
	for _, entity := range e {
		p = append(p, entity.ToProto())
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
