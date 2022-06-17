package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	stopIn = flag.Duration("stop-in", 8*time.Hour, "stop duration. eg 5s | 1m | 10h")
)

func main() {
	flag.Parse()

	timer := time.NewTimer(*stopIn)
	go func() {
		<-timer.C
		fmt.Printf("Stopping after %v at %v\n", *stopIn, time.Now())
		os.Exit(0)
	}()

	for {
		x := rand.Intn(1000)
		y := rand.Intn(1000)

		robotgo.Move(x, y)

		time.Sleep(5 * time.Second)
	}
}
