package main

import (
	"math/rand"
	"time"
)

var (
	srcRand         = rand.NewSource(time.Now().UnixNano())
	alarmIsStarted  = false
	streamIsStarted = false
)
