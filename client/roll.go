package main

import (
	"os"
	"time"
)

type RollSpeed uint

const (
	RollSpeedFast RollSpeed = iota
	RollSpeedSlow
)

var rollSpeedToMs = map[RollSpeed]time.Duration{
	RollSpeedFast: 20 * time.Millisecond,
	RollSpeedSlow: 70 * time.Millisecond,
}

func RollText(txt string, speed RollSpeed) {
	for _, c := range txt {
		time.Sleep(rollSpeedToMs[speed])
		os.Stdout.Write([]byte(string(c)))
	}
}
