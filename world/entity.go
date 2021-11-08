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

func (e *Entity) OnActions(actions []*pb.Action) {
	const speed = 4
	var velocity Vector
	for _, action := range actions {
		switch action.Action {
		case pb.Action_MOVE_UP:
			velocity.Y -= speed
		case pb.Action_MOVE_DOWN:
			velocity.Y += speed
		case pb.Action_MOVE_LEFT:
			velocity.X -= speed
		case pb.Action_MOVE_RIGHT:
			velocity.X += speed
		}
	}
	e.Velocity = velocity
}
