package main

import "testing"

func TestGPT3CostsCalculator_getNumTokens(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		ExpectedTokenUsage int
	}{
		{"Empty string", "", 0},
		{"Single word", "Hello", 1},
		{"Two words", "Hello World", 2},
		{"Many words", "Many words map to one token, but some don't: indivisible.", 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNumTokens(tt.input); got != tt.ExpectedTokenUsage {
				t.Errorf("getNumTokens() = %v, want %v", got, tt.ExpectedTokenUsage)
			}
		})
	}
}

func TestGPT3CostsCalculator_shortenToTokens(t *testing.T) {
	// function is defined like: func shortenToTokens(sentence string, tokens int) string
	tests := []struct {
		name   string
		input  string
		tokens int
		output string
	}{
		{"Empty string", "", 0, ""},
		{"Single word", "Hello", 1, "Hello"},
		{"Two words", "Hello World", 2, "Hello World"},
		{"Many words", "Many words map to one token, but some don't: indivisible.", 16, "Many words map to one token, but some don't: indivisible."},
		{"Many words", "Hello World map to one token, but some don't: indivisible.", 2, "Hello World"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortenToTokens(tt.input, tt.tokens); got != tt.output {
				t.Errorf("shortenToTokens() = %v, want %v", got, tt.output)
			}
		})
	}
}
