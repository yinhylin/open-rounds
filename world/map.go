package world

import (
	"bufio"
	"errors"
	"rounds/pb"
	"strconv"
	"strings"
)

type tileIndex int

const tileSize = 32

const (
	skyTile tileIndex = iota
	platformTile
)

var tileIndices = []Tile{
	// skyTile
	{
		Dense: false,
		Image: "sky",
	},
	// platformTile
	{
		Dense: true,
		Image: "platform",
	},
}

type Tile struct {
	Dense bool
	Image string
}

type Map struct {
	Tiles  []tileIndex
	Width  int64
	Height int64
}

func (m *Map) ToProto() *pb.Map {
	tiles := make([]int64, len(m.Tiles))
	for i, tile := range m.Tiles {
		tiles[i] = int64(tile)
	}
	return &pb.Map{
		Tiles:  tiles,
		Width:  int64(m.Width),
		Height: int64(m.Height),
	}
}

func MapFromProto(p *pb.Map) *Map {
	tiles := make([]tileIndex, len(p.Tiles))
	for i, tile := range p.Tiles {
		tiles[i] = tileIndex(tile)
	}
	return &Map{
		Tiles:  tiles,
		Width:  p.Width,
		Height: p.Height,
	}
}

func (m *Map) At(x, y int64) (*Tile, error) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil, errors.New("out of bounds")
	}
	return &tileIndices[m.Tiles[m.Width*y+x]], nil
}

func (m *Map) ForEach(callback func(x, y int64, tile Tile)) {
	for y := int64(0); y < m.Height; y++ {
		for x := int64(0); x < m.Width; x++ {
			callback(x, y, tileIndices[m.Tiles[m.Width*y+x]])
		}
	}
}

func LoadMap(contents string) (*Map, error) {
	scanner := bufio.NewScanner(strings.NewReader(contents))

	scanner.Scan()
	width, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, err
	}

	scanner.Scan()
	height, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, err
	}

	tiles := make([]tileIndex, 0, width*height)
	for scanner.Scan() {
		for _, item := range scanner.Text() {
			switch item {
			case '.':
				tiles = append(tiles, skyTile)
			case '#':
				tiles = append(tiles, platformTile)
			}
		}
	}

	return &Map{
		Tiles:  tiles,
		Width:  int64(width),
		Height: int64(height),
	}, nil
}
