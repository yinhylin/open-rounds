package utils

import (
	"log"
	"regexp"
	"testing"
)

// TestReadToml calls ReadToml with a known test config, checking
// for a valid return value for each key
func TestReadToml(t *testing.T) {
	ReadToml("testConf.toml")
	var wantRegex = regexp.MustCompile("test")
	var barVal = Cfg.Game.Bar

	log.Printf("config: %v", Cfg)

	if !wantRegex.MatchString(barVal) {
		t.Fatalf(`bar = %q, want match for %#q`, barVal, wantRegex)
	}

	// var wantFloat = float32(1)
	var wantFloat = 1
	var jumpHeightVal = Cfg.Player.JumpHeight
	if wantFloat != jumpHeightVal {
		t.Fatalf(`Player.jumpHeight = %v, want match for %#v`, jumpHeightVal, wantFloat)
	}

	// wantFloat = float32(1)
	wantFloat = 1
	var speedVal = Cfg.Player.Speed
	if wantFloat != speedVal {
		t.Fatalf(`Player.speed = %v, want match for %#v`, speedVal, wantFloat)
	}

	var wantInt = 1
	var resXval = Cfg.Ui.Resolution.X
	if wantInt != resXval {
		t.Fatalf(`UI.Resolution.X = %v, want match for %#v`, resXval, wantInt)
	}

	wantInt = 1
	var resYval = Cfg.Ui.Resolution.Y
	if wantInt != resYval {
		t.Fatalf(`UI.Resolution.Y = %v, want match for %#v`, resYval, wantInt)
	}
}
