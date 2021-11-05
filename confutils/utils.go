package confutils

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type PlayerConfig struct {
	speed      float32
	jumpHeight float32
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

type Config struct {
	Player PlayerConfig
	Ui     UIConfig
	Game   GameConfig
}

func ReadToml(fileName string) Config {
	var cfg Config
	cfg_file, file_err := os.ReadFile(fileName)
	if file_err != nil {
		panic(file_err)
	}
	log.Printf("config file: %s", cfg_file)

	toml_err := toml.Unmarshal([]byte(cfg_file), &cfg)
	if toml_err != nil {
		panic(toml_err)
	}

	log.Printf("config: %v", cfg)
	log.Printf("config.UI: %v", cfg.Ui)

	return cfg
}
