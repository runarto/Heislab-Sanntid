package main

import (
	"math"
)

func AbsValue(x int, y int) int { 
	return int(math.Abs(float64(x - y)))
}