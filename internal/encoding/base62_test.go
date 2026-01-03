package encoding

import "testing"

func TestEncode(t *testing.T) {
	// A slice of structs makes it easy to define multiple test cases (inputs and expected outputs).
	tests := []struct {
		input    uint64
		expected string
	}{
		{input: 0, expected: "0"},
		{input: 10, expected: "A"},
		{input: 61, expected: "z"}, // Max single character
		{input: 62, expected: "10"}, // Min two characters
	}

	for _, test := range tests {
		actual := Encode(test.input)
		
		if actual != test.expected {
			// t.Errorf prints the error message and marks the test as failed.
			t.Errorf("Encode(%d) failed: Expected '%s', Got '%s'", test.input, test.expected, actual)
		}
	}
}