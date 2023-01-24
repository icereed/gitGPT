package main

import "testing"

func Test_cleanString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"!!Hello World!!", "Hello World!!"},
		{"!!He!!llo World", "He!!llo World"},
		{"Hello World!!", "Hello World!!"},
		{"Hello World", "Hello World"},
		{"1Hello World!", "Hello World!"},
		{"###########\nHello World! \n ##########", "Hello World!"},
		{"###########\nHello World. \n ##########", "Hello World."},
	}

	for _, test := range testCases {
		output := cleanString(test.input)
		if output != test.expected {
			t.Errorf("cleanString(%s) = %s, expected %s", test.input, output, test.expected)
		}
	}
}
