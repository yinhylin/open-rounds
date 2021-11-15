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
	var jumpHeightVal = config.Player.JumpHeight
	if !AlmostEqual(wantFloat, jumpHeightVal, threshold) {
		t.Fatalf(`Player.jumpHeight = %v, want match for %#v, diff %v with threshold %v`, jumpHeightVal, wantFloat, wantFloat-jumpHeightVal, threshold)
	}

	wantFloat = float64(1.0)
	var speedVal = config.Player.Speed
	if !AlmostEqual(wantFloat, speedVal, threshold) {
		t.Fatalf(`Player.speed = %v, want match for %#v, diff %v with threshold %v`, speedVal, wantFloat, wantFloat-speedVal, threshold)
	}

	var wantInt = 1
	var resXval = config.UI.Resolution.X
	if wantInt != resXval {
		t.Fatalf(`UI.Resolution.X = %v, want match for %#v`, resXval, wantInt)
	}

	wantInt = 1
	var resYval = config.UI.Resolution.Y
	if wantInt != resYval {
		t.Fatalf(`UI.Resolution.Y = %v, want match for %#v`, resYval, wantInt)
	}
}

type LerpInput struct {
	Start, End, T float64
	Expected float64
}

func TestLerp(t *testing.T) {
	var testVals = [...]LerpInput{
		{0, 100, 0.5, 50},
		{20, 80, 0, 20},
		{-1, 1, 0.5, 0},
		{0.5, 1, 0.5, 0.75},
	}
	for _, aTest := range testVals {
		if Lerp(aTest.Start, aTest.End, aTest.T) != aTest.Expected {
			t.Fatalf(`Start = %v, End = %v, T = %v, expected %v but got %v`,
				aTest.Start, aTest.End, aTest.T, aTest.Expected, Lerp(aTest.Start, aTest.End, aTest.T))
		}
	}
}
