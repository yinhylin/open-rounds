package world

import "rounds/pb"

type Vector struct {
	X, Y float64
}

// TODO: This all needs a giant refactor to dedupe things. :(
type Entity struct {
	ID       string
	Coords   Vector
	Velocity Vector
	Intents  map[pb.Intents_Intent]struct{}
	Angle    float64
}

type Bullet struct {
	ID       string
	Coords   Vector
	Velocity Vector
	Angle    float64
}

func (v *Vector) ToProto() *pb.Vector {
	return &pb.Vector{
		X: v.X,
		Y: v.Y,
	}
}

func VectorFromProto(v *pb.Vector) Vector {
	return Vector{
		X: v.X,
		Y: v.Y,
	}
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

func (e *Entity) ToProto() *pb.Entity {
	return &pb.Entity{
		Id:       e.ID,
		Position: e.Coords.ToProto(),
		Velocity: e.Velocity.ToProto(),
		Intents:  IntentsToProto(e.Intents),
		Angle:    e.Angle,
	}
}

func EntityFromProto(e *pb.Entity) *Entity {
	if e == nil {
		return nil
	}
	return &Entity{
		ID:       e.Id,
		Coords:   VectorFromProto(e.Position),
		Velocity: VectorFromProto(e.Velocity),
		Intents:  IntentsFromProto(e.Intents),
		Angle:    e.Angle,
	}
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
