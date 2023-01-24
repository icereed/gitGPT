package main

import "testing"

func TestGPT3CostsCalculator_getNumTokens(t *testing.T) {
	type args struct {
		sentence string
	}
	tests := []struct {
		name               string
		input              string
		ExpectedTokenUsage int
	}{
		{"Empty string", "", 0},
		{"Single word", "Hello", 2},
		{"Two words", "Hello World", 4},
		{"Many words", "Many words map to one token, but some don't: indivisible.", 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, _ := NewGPT3CostsCalculator("davinci")
			if got := g.getNumTokens(tt.input); got != tt.ExpectedTokenUsage {
				t.Errorf("GPT3CostsCalculator.getNumTokens() = %v, want %v", got, tt.ExpectedTokenUsage)
			}
		})
	}
}
