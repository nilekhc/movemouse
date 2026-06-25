package main

import "testing"

func TestShouldNudge(t *testing.T) {
	cases := []struct {
		name      string
		lastKnown Point
		current   Point
		want      bool
	}{
		{name: "equal positions → nudge", lastKnown: Point{X: 100, Y: 200}, current: Point{X: 100, Y: 200}, want: true},
		{name: "differing positions → yield", lastKnown: Point{X: 100, Y: 200}, current: Point{X: 101, Y: 200}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldNudge(tc.lastKnown, tc.current)
			if got != tc.want {
				t.Errorf("shouldNudge(%v, %v) = %v, want %v", tc.lastKnown, tc.current, got, tc.want)
			}
		})
	}
}
