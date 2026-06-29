package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// box geometry. all rows render to exactly boxWidth runes wide,
// with the side borders consuming 2 runes and a 1-rune padding
// on each side reserved for legibility.
const (
	boxWidth          = 45
	innerWidth        = boxWidth - 2   // between the two │ borders
	innerContentWidth = innerWidth - 2 // between the "│ " and " │" gutters
	barWidth          = 20
	labelWidth        = 9 // widest label is "remaining"
)

// ansi sequences used to redraw the box in place. cursorUp7 moves the
// cursor back to the top of a 7-line frame (top border + 5 rows + bottom
// border). each row is printed with a trailing newline, so after a
// 7-line frame the cursor sits on line 8 — we move up 7 to land on
// line 1. the done frame is 8 lines (adds a "done" row before the
// bottom border); it still uses cursorUp7 because it is only ever
// drawn from line 8 — the extra 8th line claims a fresh row below.
const (
	cursorUp7 = "\033[7A"
	clearLine = "\r\033[K"
)

// dashboard renders the 7-line status panel for movemouse. it owns
// no goroutines — the caller drives redraws by calling render() on a
// 1s cadence, and done() once on context cancellation.
type dashboard struct {
	startedAt time.Time
	stopIn    time.Duration
}

// newDashboard returns a dashboard bound to a start time and a stop
// deadline. the deadline is used to compute the ends-at clock and the
// progress ratio.
func newDashboard(startedAt time.Time, stopIn time.Duration) *dashboard {
	return &dashboard{startedAt: startedAt, stopIn: stopIn}
}

// start prints the initial 7-line frame at the current cursor position.
// must be called exactly once, before any render() or done() calls.
func (d *dashboard) start() {
	d.print(d.liveRows(d.startedAt, 0))
}

// render rewrites all 7 live-frame rows in place. it computes elapsed
// and remaining from now and the deadline, clamping remaining at zero
// (so we never show a negative countdown if a redraw lands just past
// the deadline).
func (d *dashboard) render(now time.Time) {
	fmt.Print(cursorUp7)
	d.print(d.liveRows(now, d.elapsed(now)))
}

// done emits a final 8-line frame: the live rows plus a "done" row
// spliced in just before the bottom border, so the run summary lives
// inside the box rather than below it. used for both natural -stop-in
// expiry and SIGINT/SIGTERM — same final-frame format in either case.
// the bar reflects the real ratio at exit (so an early ctrl-c shows
// partial progress, not a misleading 100%).
func (d *dashboard) done(endedAt time.Time) {
	fmt.Print(cursorUp7)
	elapsed := d.elapsed(endedAt)
	rows := d.liveRows(endedAt, elapsed)
	// splice the done row in just before the bottom border (last row).
	doneValue := fmt.Sprintf("%s, ran for %s",
		endedAt.Format("3:04:05 PM"), elapsed.Truncate(time.Second))
	rows = append(rows[:len(rows)-1], kvRow("done", doneValue), rows[len(rows)-1])
	d.print(rows)
}

// elapsed returns the time between startedAt and now, never negative.
func (d *dashboard) elapsed(now time.Time) time.Duration {
	return max(0, now.Sub(d.startedAt))
}

// liveRows builds the 7-line live frame (top border + 5 kv rows +
// bottom border). elapsed drives both the elapsed row and the progress
// ratio; the ratio is clamped to [0, 1] so a redraw landing past the
// deadline still shows a sane 100% bar rather than overflowing.
func (d *dashboard) liveRows(now time.Time, elapsed time.Duration) []string {
	ratio := float64(elapsed) / float64(d.stopIn)
	if ratio > 1.0 {
		ratio = 1.0
	}
	if ratio < 0 {
		ratio = 0
	}
	remaining := max(0, d.stopIn-elapsed)
	endsAt := d.startedAt.Add(d.stopIn)
	_ = now // present for future use; rows derive everything from elapsed/ratio
	return []string{
		topBorder("movemouse"),
		kvRow("started", d.startedAt.Format("3:04:05 PM")),
		kvRow("ends", endsAt.Format("3:04:05 PM")),
		kvRow("elapsed", elapsed.Truncate(time.Second).String()),
		kvRow("remaining", remaining.Truncate(time.Second).String()),
		progressRow(ratio),
		bottomBorder(),
	}
}

// print writes every row to stdout with a clear-line prefix and a
// trailing newline. callers are responsible for positioning the cursor
// (e.g. cursorUp7) before the first call on subsequent redraws.
func (d *dashboard) print(rows []string) {
	for _, row := range rows {
		fmt.Print(clearLine + row + "\n")
	}
}

// topBorder returns the title border row: ┌─ <title> ───...───┐ padded
// to exactly boxWidth runes.
func topBorder(title string) string {
	// "┌─ " + title + " " consumes 4 + len(title) runes (the box characters
	// each count as 1 rune even though they're multi-byte in UTF-8).
	prefix := "┌─ " + title + " "
	pad := max(0, boxWidth-runeLen(prefix)-1) // -1 for trailing ┐
	return prefix + strings.Repeat("─", pad) + "┐"
}

// bottomBorder returns └───...───┘ of exactly boxWidth runes.
func bottomBorder() string {
	return "└" + strings.Repeat("─", innerWidth) + "┘"
}

// kvRow returns │ <label-padded> <value-padded-to-fill> │ of exactly
// boxWidth runes. label is left-padded to labelWidth so values column-
// align across rows.
func kvRow(label, value string) string {
	content := fmt.Sprintf("%-*s %s", labelWidth, label, value)
	return wrapRow(content)
}

// progressRow returns the progress row: a kvRow whose value is
// "[<bar>]  <pct>%" with the percent right-aligned in 6 runes so it
// reads cleanly from "  0.0%" through "100.0%".
func progressRow(ratio float64) string {
	pct := fmt.Sprintf("%5.1f%%", ratio*100)
	value := fmt.Sprintf("[%s]  %s", bar(ratio, barWidth), pct)
	return kvRow("progress", value)
}

// bar returns width cells: round(ratio*width) filled with █, the rest
// with ░. ratio is clamped to [0, 1] before rounding so callers can
// pass slightly-out-of-range values without bar producing garbage.
func bar(ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := min(width, int(math.Round(ratio*float64(width))))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// wrapRow returns │ <content padded to innerContentWidth> │ of exactly
// boxWidth runes. content longer than innerContentWidth is truncated.
func wrapRow(content string) string {
	if runeLen(content) > innerContentWidth {
		// truncate by runes, not bytes, to keep unicode safe.
		runes := []rune(content)
		content = string(runes[:innerContentWidth])
	}
	pad := innerContentWidth - runeLen(content)
	return "│ " + content + strings.Repeat(" ", pad) + " │"
}

// runeLen returns the visual rune count of s. used because box-drawing
// characters and emoji are multi-byte in UTF-8 — len(s) would over-count
// and break the padding math.
func runeLen(s string) int {
	return len([]rune(s))
}
