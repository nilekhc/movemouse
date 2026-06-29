package main

import (
	"strings"
	"testing"
)

func TestTopBorder(t *testing.T) {
	got := topBorder("movemouse")
	want := "┌─ movemouse " + strings.Repeat("─", 31) + "┐"
	if got != want {
		t.Errorf("topBorder(%q) = %q, want %q", "movemouse", got, want)
	}
	if runeLen(got) != boxWidth {
		t.Errorf("topBorder rune width = %d, want %d", runeLen(got), boxWidth)
	}
}

func TestBottomBorder(t *testing.T) {
	got := bottomBorder()
	want := "└" + strings.Repeat("─", innerWidth) + "┘"
	if got != want {
		t.Errorf("bottomBorder() = %q, want %q", got, want)
	}
	if runeLen(got) != boxWidth {
		t.Errorf("bottomBorder rune width = %d, want %d", runeLen(got), boxWidth)
	}
}

func TestKvRow(t *testing.T) {
	cases := []struct {
		name  string
		label string
		value string
		want  string
	}{
		{
			name:  "short value",
			label: "started",
			value: "4:28:53 PM",
			// content = "started  " (9-wide) + " " + "4:28:53 PM" = 20 runes; pad to 41 = 21 spaces.
			want: "│ started   4:28:53 PM" + strings.Repeat(" ", 21) + " │",
		},
		{
			name:  "longest label aligns",
			label: "remaining",
			value: "1h29m30s",
			// content = "remaining" (9) + " " + "1h29m30s" (8) = 18 runes; pad = 23.
			want: "│ remaining 1h29m30s" + strings.Repeat(" ", 23) + " │",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := kvRow(tc.label, tc.value)
			if got != tc.want {
				t.Errorf("kvRow(%q, %q) = %q, want %q", tc.label, tc.value, got, tc.want)
			}
			if runeLen(got) != boxWidth {
				t.Errorf("kvRow rune width = %d, want %d", runeLen(got), boxWidth)
			}
		})
	}
}

func TestProgressRow(t *testing.T) {
	cases := []struct {
		name    string
		ratio   float64
		wantBar string
		wantPct string
	}{
		{name: "zero", ratio: 0.0, wantBar: strings.Repeat("░", 20), wantPct: "  0.0%"},
		{name: "half", ratio: 0.5, wantBar: strings.Repeat("█", 10) + strings.Repeat("░", 10), wantPct: " 50.0%"},
		{name: "full", ratio: 1.0, wantBar: strings.Repeat("█", 20), wantPct: "100.0%"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := progressRow(tc.ratio)
			// value within the row: "[" + bar + "]  " + pct = 1 + 20 + 3 + 6 = 30 runes.
			// kvRow content: "progress " (9) + " " + value (30) = 40 runes; pad = 1.
			want := "│ progress  [" + tc.wantBar + "]  " + tc.wantPct + " " + " │"
			if got != want {
				t.Errorf("progressRow(%v) =\n  got: %q\n want: %q", tc.ratio, got, want)
			}
			if runeLen(got) != boxWidth {
				t.Errorf("progressRow rune width = %d, want %d", runeLen(got), boxWidth)
			}
		})
	}
}

func TestBar(t *testing.T) {
	cases := []struct {
		name  string
		ratio float64
		width int
		want  string
	}{
		{name: "empty", ratio: 0.0, width: 20, want: strings.Repeat("░", 20)},
		{name: "half", ratio: 0.5, width: 20, want: strings.Repeat("█", 10) + strings.Repeat("░", 10)},
		{name: "full", ratio: 1.0, width: 20, want: strings.Repeat("█", 20)},
		{name: "rounds up at 0.06×20=1.2", ratio: 0.06, width: 20, want: "█" + strings.Repeat("░", 19)},
		{name: "negative ratio clamped to 0", ratio: -0.5, width: 20, want: strings.Repeat("░", 20)},
		{name: "above 1 clamped to full", ratio: 1.5, width: 20, want: strings.Repeat("█", 20)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := bar(tc.ratio, tc.width)
			if got != tc.want {
				t.Errorf("bar(%v, %d) = %q, want %q", tc.ratio, tc.width, got, tc.want)
			}
			if runeLen(got) != tc.width {
				t.Errorf("bar rune width = %d, want %d", runeLen(got), tc.width)
			}
		})
	}
}
