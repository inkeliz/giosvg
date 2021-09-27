package svgparser

import (
	"math"
	"testing"
)

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
func TestUnits(t *testing.T) {
	data := []struct {
		s   string
		val float64
	}{
		{s: "1.4%", val: 1.4},
		{s: "1.", val: 1.},
		{s: "1.px", val: 1.},
		{s: "1.2", val: 1.2},
		{s: "1.2pt", val: 1.6},
		{s: "10 px", val: 10},
		{s: "10 cm", val: 377.9527559055},
		{s: "10 mm", val: 37.7952755906},
		{s: "10 in", val: 960},
		{s: "10 pt", val: 13.3333333333},
		{s: "10 pc", val: 160},
	}
	for i, d := range data {
		value, isPerc, err := parseUnit(d.s)
		if err != nil {
			t.Fatal(err)
		}
		if isPerc != (i == 0) {
			t.Fatalf("%s is not a percentage", d.s)
		}
		if !almostEqual(value, d.val) {
			t.Fatalf("for %s, expected %.10f, got %.10f", d.s, d.val, value)
		}
	}
}
