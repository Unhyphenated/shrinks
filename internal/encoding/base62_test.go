package encoding

import (
	"math"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{input: 0, expected: "0"},
		{input: 10, expected: "A"},
		{input: 61, expected: "z"},
		{input: 62, expected: "10"},
	}

	for _, test := range tests {
		actual := Encode(test.input)
		if actual != test.expected {
			t.Errorf("Encode(%d) = %s, want %s", test.input, actual, test.expected)
		}
	}
}

// ===== NEW TESTS =====

// Test #4: Large numbers encode correctly
func TestEncode_LargeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"one million", 1_000_000, "4C92"},
		{"one billion", 1_000_000_000, "15ftgG"},
		{"realistic DB id", 123456789, "8M0kX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Encode(tt.input)
			if actual != tt.expected {
				t.Errorf("Encode(%d) = %s, want %s", tt.input, actual, tt.expected)
			}
		})
	}
}

// Test #5: Max uint64 doesn't overflow or panic
func TestEncode_MaxUint64(t *testing.T) {
	maxUint := uint64(math.MaxUint64)

	// Should not panic
	result := Encode(maxUint)

	// Should produce a non-empty string
	if result == "" {
		t.Error("Encode(MaxUint64) returned empty string")
	}

	// MaxUint64 in base62 should be "LygHa16AHYF" (11 chars)
	expectedLen := 11
	if len(result) != expectedLen {
		t.Errorf("Encode(MaxUint64) length = %d, want %d", len(result), expectedLen)
	}
}

// Test #6: Valid codes decode correctly
func TestDecode_ValidCode(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0", 0},
		{"A", 10},
		{"z", 61},
		{"10", 62},
		{"4C92", 1_000_000},
		{"8M0kX", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual, err := Decode(tt.input)
			if err != nil {
				t.Fatalf("Decode(%s) returned error: %v", tt.input, err)
			}
			if actual != tt.expected {
				t.Errorf("Decode(%s) = %d, want %d", tt.input, actual, tt.expected)
			}
		})
	}
}

// Test #7: Invalid characters return error
func TestDecode_InvalidChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"hyphen", "abc-def"},
		{"underscore", "abc_def"},
		{"space", "abc def"},
		{"special char", "abc!def"},
		{"unicode", "abc日def"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.input)
			if err == nil {
				t.Errorf("Decode(%s) should return error for invalid input", tt.input)
			}
		})
	}
}

// Test #8: Encode then Decode returns original value
func TestEncodeDecode_RoundTrip(t *testing.T) {
	testValues := []uint64{
		0,
		1,
		61,
		62,
		100,
		1000,
		123456789,
		1_000_000_000,
		math.MaxUint64,
	}

	for _, original := range testValues {
		encoded := Encode(original)
		decoded, err := Decode(encoded)

		if err != nil {
			t.Errorf("RoundTrip(%d): Decode returned error: %v", original, err)
			continue
		}

		if decoded != original {
			t.Errorf("RoundTrip(%d): Encode=%s, Decode=%d, want %d",
				original, encoded, decoded, original)
		}
	}
}

// Bonus: Ensure encoded strings only contain valid base62 characters
func TestEncode_OnlyValidChars(t *testing.T) {
	testValues := []uint64{0, 1, 62, 1000, 123456789, math.MaxUint64}

	for _, val := range testValues {
		encoded := Encode(val)
		for i, char := range encoded {
			idx := indexOfChar(byte(char))
			if idx == -1 {
				t.Errorf("Encode(%d) = %s, contains invalid char '%c' at position %d",
					val, encoded, char, i)
			}
		}
	}
}

// Helper: find index of character in alphabet (returns -1 if not found)
func indexOfChar(c byte) int {
	for i := 0; i < len(Alphabet); i++ {
		if Alphabet[i] == c {
			return i
		}
	}
	return -1
}
