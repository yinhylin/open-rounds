package utils

import (
	"log"
	"math"
	"regexp"
	"testing"
)

type AlmostEqualInput struct {
	A, B, threshold float64
	IsAlmostEqual   bool
}

// TestAlmostEqual calls AlmostEqual with known floats, checking
// if the threshhold is workign as expected
func TestAlmostEqual(t *testing.T) {
	var testVals = [...](AlmostEqualInput){
		AlmostEqualInput{1.0, 1.0, 1.0e-9, true},
		AlmostEqualInput{1.0 + 1e-10, 1.0, 1.0e-9, true},
		AlmostEqualInput{1.0, 1.0 + 1e-10, 1.0e-9, true},
		AlmostEqualInput{1.0, 1.0 + 1e-6, 1.0e-9, false},
		AlmostEqualInput{1.0 + 1e-6, 1.0, 1.0e-9, false},
	}
	//for each x,y in testvals, check comparison
	for _, aTest := range testVals {
		if AlmostEqual(aTest.A, aTest.B, aTest.threshold) != aTest.IsAlmostEqual {
			t.Fatalf(`A = %v, B=%v, threshold = %v, expect isAlmostEqual == %v but got %v . Diff == %v`,
				aTest.A, aTest.B, aTest.threshold, aTest.IsAlmostEqual, AlmostEqual(aTest.A, aTest.B, aTest.threshold), math.Abs(aTest.A-aTest.B))
		}
	}
}

// TestReadToml calls ReadToml with a known test config, checking
// for a valid return value for each key
func TestReadToml(t *testing.T) {
	ReadToml("testConf.toml")
	var wantRegex = regexp.MustCompile("test")
	var barVal = Cfg.Game.Bar

	log.Printf("config: %v", Cfg)

	var wantFloat = float64(1.0)
	var threshold = Cfg.Math.Float64EqualityThreshold
	if AlmostEqual(wantFloat, threshold, 1e-09) {
		t.Fatalf(`threshold = %v, want match for %#v, diff %v`, threshold, wantFloat, wantFloat-threshold)
	}

	if !wantRegex.MatchString(barVal) {
		t.Fatalf(`bar = %q, want match for %#q`, barVal, wantRegex)
	}

	wantFloat = float64(1.0)
	var jumpHeightVal = Cfg.Player.JumpHeight
	if !AlmostEqual(wantFloat, jumpHeightVal, threshold) {
		t.Fatalf(`Player.jumpHeight = %v, want match for %#v, diff %v with threshold %v`, jumpHeightVal, wantFloat, wantFloat-jumpHeightVal, threshold)
	}

	wantFloat = float64(1.0)
	var speedVal = Cfg.Player.Speed
	if !AlmostEqual(wantFloat, speedVal, threshold) {
		t.Fatalf(`Player.speed = %v, want match for %#v, diff %v with threshold %v`, speedVal, wantFloat, wantFloat-speedVal, threshold)
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
