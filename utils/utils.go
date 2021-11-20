package utils

import (
	"math"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type PlayerConfig struct {
	Speed, JumpHeight float64
}

type GameConfig struct {
	Bar string
}

type ResolutionConfig struct {
	X, Y int
}

type UIConfig struct {
	Resolution ResolutionConfig
}

type MathConfig struct {
	Float64EqualityThreshold float64
}

type Config struct {
	Player PlayerConfig
	UI     UIConfig
	Game   GameConfig
	Math   MathConfig
}

func ReadTOML(fileName string) (*Config, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := toml.Unmarshal([]byte(file), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func AlmostEqual(a, b, threshold float64) bool {
	return math.Abs(a-b) <= threshold
}
