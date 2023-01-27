package main

import (
	"os/user"
	"testing"
)

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
		{"Hello World...", "Hello World..."},
	}

	for _, test := range testCases {
		output := cleanString(test.input)
		if output != test.expected {
			t.Errorf("cleanString(%s) = %s, expected %s", test.input, output, test.expected)
		}
	}
}

func Test_expandPath(t *testing.T) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	type args struct {
		pathToBeExpanded string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"No expansion", args{"/bla/blub"}, "/bla/blub"},
		{"Home expansion", args{"~/Hello World"}, dir + "/Hello World"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.args.pathToBeExpanded)
			if got != tt.want {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
