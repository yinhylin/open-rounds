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
	Ui     UIConfig
	Game   GameConfig
	Math   MathConfig
}

var Cfg Config

func ReadToml(fileName string) Config {
	cfg_file, file_err := os.ReadFile(fileName)
	if file_err != nil {
		panic(file_err)
	}

	toml_err := toml.Unmarshal([]byte(cfg_file), &Cfg)
	if toml_err != nil {
		panic(toml_err)
	}

	return Cfg
}

func AlmostEqual(a, b, threshold float64) bool {
	return math.Abs(a-b) <= threshold
}
