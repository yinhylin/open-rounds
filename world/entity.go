package world

import "rounds/pb"

type Vector struct {
	X, Y float64
}

type Entity struct {
	ID       string
	Coords   Vector
	Velocity Vector
	Intents  map[pb.Intents_Intent]struct{}
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
