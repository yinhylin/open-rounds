package world

import "rounds/pb"

type Vector struct {
	X, Y float64
}

type Entity struct {
	ID       string
	Coords   Vector
	Velocity Vector
	Actions  map[pb.Actions_Event]struct{}
}

func (e *Entity) Update() {
	const speed = 4

	// TODO: Don't overwrite velocity.
	var velocity Vector
	for action := range e.Actions {
		switch action {
		case pb.Actions_MOVE_UP:
			velocity.Y -= speed
		case pb.Actions_MOVE_DOWN:
			velocity.Y += speed
		case pb.Actions_MOVE_LEFT:
			velocity.X -= speed
		case pb.Actions_MOVE_RIGHT:
			velocity.X += speed
		}
	}

	// TODO: This needs to move elsewhere and have collision checking.
	e.Velocity = velocity
	e.Coords.X += e.Velocity.X
	e.Coords.Y += e.Velocity.Y
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

func ActionsToProto(actions map[pb.Actions_Event]struct{}) *pb.Actions {
	var events []pb.Actions_Event
	for action := range actions {
		events = append(events, action)
	}
	return &pb.Actions{
		Actions: events,
	}
}

func ActionsFromProto(a *pb.Actions) map[pb.Actions_Event]struct{} {
	actions := make(map[pb.Actions_Event]struct{})
	for _, action := range a.Actions {
		actions[action] = struct{}{}
	}
	return actions
}

func (e *Entity) ToProto() *pb.Entity {
	return &pb.Entity{
		Id:       e.ID,
		Position: e.Coords.ToProto(),
		Velocity: e.Velocity.ToProto(),
		Actions:  ActionsToProto(e.Actions),
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
		Actions:  ActionsFromProto(e.Actions),
	}
}
