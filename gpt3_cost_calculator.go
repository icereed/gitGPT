package main

import (
	"math"
	"strings"
)

func getNumTokens(sentence string) int {
	return int(math.Ceil((float64(len(sentence)) - float64(strings.Count(sentence, " "))) / 1.4))
}

func getNumChars(numTokens int) int {
	return int(math.Floor(float64(numTokens) * 1.5))
}
