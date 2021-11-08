package world

import "rounds/pb"

type Coords struct {
	X, Y float64
}

func (c *Coords) Coordinates() Coords {
	return *c
}

type Vector struct {
	X, Y float64
}

type Entity struct {
	Coords
	Velocity Vector
}

func (e *Entity) Update() {
	e.X += e.Velocity.X
	e.Y += e.Velocity.Y
}

func (e *Entity) OnAction(action *pb.Action) {
	const speed = 4
	switch action.Action {
	case pb.Action_MOVE_UP:
		e.Velocity.Y = -speed
	case pb.Action_MOVE_DOWN:
		e.Velocity.Y = speed
	case pb.Action_MOVE_LEFT:
		e.Velocity.X = -speed
	case pb.Action_MOVE_RIGHT:
		e.Velocity.X = speed
	case pb.Action_NONE:
		e.Velocity.X = 0
		e.Velocity.Y = 0
	}
}
