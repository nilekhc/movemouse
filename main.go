package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	stopIn   = flag.Duration("stop-in", 8*time.Hour, "stop duration. eg 5s | 1m | 10h")
	interval = flag.Duration("interval", 30*time.Second, "tick cadence between nudges. eg 10s | 1m")
	jitter   = flag.Int("jitter", 3, "nudge delta in pixels (uniform random in [-jitter, +jitter])")
)

// Point is a cursor coordinate. compared by value via ==.
type Point struct{ X, Y int }

// shouldNudge reports whether movemouse should nudge this tick. it returns true
// only when the cursor has not moved since the last tick — i.e. the user has not
// touched the mouse. when they have, we yield and re-baseline on the next call.
func shouldNudge(lastKnown, current Point) bool {
	return lastKnown == current
}

// pickDelta returns a uniform random (dx, dy) in [-jitter, +jitter], excluding (0, 0)
// so the nudge always actually moves the cursor.
func pickDelta(jitter int) (int, int) {
	for {
		dx := rand.Intn(2*jitter+1) - jitter
		dy := rand.Intn(2*jitter+1) - jitter
		if dx != 0 || dy != 0 {
			return dx, dy
		}
	}
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	timer := time.AfterFunc(*stopIn, cancel)
	defer timer.Stop()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	x, y := robotgo.Location()
	lastKnown := Point{X: x, Y: y}

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("stopping at %v\n", time.Now())
			return
		case <-ticker.C:
			cx, cy := robotgo.Location()
			current := Point{X: cx, Y: cy}
			if !shouldNudge(lastKnown, current) {
				// user moved the cursor — yield and re-baseline.
				lastKnown = current
				continue
			}
			dx, dy := pickDelta(*jitter)
			nx, ny := current.X+dx, current.Y+dy
			robotgo.Move(nx, ny)
			lastKnown = Point{X: nx, Y: ny}
		}
	}
}
