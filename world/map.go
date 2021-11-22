package world

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

type tileIndex int

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
	Width  int
	Height int
}

func (m *Map) At(x, y int) (*Tile, error) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil, errors.New("out of bounds")
	}
	return &tileIndices[m.Tiles[m.Width*y+x]], nil
}

func (m *Map) ForEach(callback func(x, y int, tile Tile)) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
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
		Width:  width,
		Height: height,
	}, nil
}
