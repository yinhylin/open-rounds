package world

type Coords struct {
	X, Y float64
}

func (c *Coords) Coordinates() Coords {
	return *c
}

type Vector struct {
	DX, DY float64
}

type Entity struct {
	Coords
}
