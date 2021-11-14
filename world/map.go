package world

type Tile struct {
	Dense  bool
	Coords Vector
	Image  string
}

type Map struct {
	Tiles []Tile
}
