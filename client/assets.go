package client

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
	"path/filepath"
	"rounds/world"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	dir = "assets"
)

//go:embed assets/*
var assets embed.FS

//go:embed assets/version.txt
var Version string

type Assets struct {
	images map[string]*ebiten.Image
	maps   map[string]*world.Map
}

func (a *Assets) Image(name string) *ebiten.Image {
	image := a.images[name]
	if image == nil {
		log.Fatalf("invalid image name: %s", name)
	}
	return image
}

func (a *Assets) Map(name string) *world.Map {
	m := a.maps[name]
	if m == nil {
		log.Fatalf("invalid map name: %s", name)
	}
	return m
}

func LoadAssets() (*Assets, error) {
	a := &Assets{
		images: make(map[string]*ebiten.Image),
		maps:   make(map[string]*world.Map),
	}

	files, err := assets.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			// yolo no recursion yet
			continue
		}

		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		switch filepath.Ext(strings.ToLower(f.Name())) {
		case ".png":
			if _, ok := a.images[name]; ok {
				return nil, fmt.Errorf("duplicate filename: %s", name)
			}

			// Can't use filepath.Join due to Windows using backlash and assets expecting a forward slash.
			file, err := assets.Open(strings.Join([]string{dir, f.Name()}, "/"))
			if err != nil {
				return nil, err
			}
			decoded, _, err := image.Decode(file)
			if err != nil {
				return nil, err
			}
			a.images[name] = ebiten.NewImageFromImage(decoded)
		case ".map":
			if _, ok := a.maps[name]; ok {
				return nil, fmt.Errorf("duplicate filename: %s", name)
			}
			file, err := assets.Open(strings.Join([]string{dir, f.Name()}, "/"))
			if err != nil {
				return nil, err
			}
			contents, err := ioutil.ReadAll(file)
			if err != nil {
				return nil, err
			}
			m, err := world.LoadMap(string(contents))
			if err != nil {
				return nil, err
			}
			a.maps[name] = m
		}
	}
	return a, nil
}
