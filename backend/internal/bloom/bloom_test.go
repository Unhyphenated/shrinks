//go:build unit
package bloom

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

// Test: Constructor computes correct m and k for known inputs
func TestNewBloomFilter_CorrectSizing(t *testing.T) {
	f := NewBloomFilter(1_000_000, 0.01)

	// m = -(n * ln(p)) / (ln(2))^2
	expectedM := uint64(math.Ceil(-(1_000_000 * math.Log(0.01)) / (math.Ln2 * math.Ln2)))
	// k = (m / n) * ln(2)
	expectedK := uint64(math.Ceil(float64(expectedM) / 1_000_000 * math.Ln2))

	if f.bitCount != expectedM {
		t.Errorf("bitCount = %d, want %d", f.bitCount, expectedM)
	}

	if f.hashCount != expectedK {
		t.Errorf("hashCount = %d, want %d", f.hashCount, expectedK)
	}

	if f.itemCount != 0 {
		t.Errorf("itemCount = %d, want 0", f.itemCount)
	}

	// Verify byte slice is correctly sized
	expectedBytes := (expectedM + 7) / 8
	if uint64(len(f.bitSet)) != expectedBytes {
		t.Errorf("bitSet length = %d bytes, want %d", len(f.bitSet), expectedBytes)
	}
}

// Test: Constructor handles edge cases without panicking
func TestNewBloomFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		n    uint64
		p    float64
	}{
		{"zero items", 0, 0.01},
		{"negative FPR", 1000, -0.5},
		{"FPR of 1.0", 1000, 1.0},
		{"FPR greater than 1", 1000, 2.0},
		{"very small FPR", 1000, 0.0001},
		{"very large item count", 100_000_000, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewBloomFilter(tt.n, tt.p)
			if f == nil {
				t.Fatal("NewBloomFilter returned nil")
			}
			if f.bitCount == 0 {
				t.Error("bitCount should be > 0")
			}
			if f.hashCount == 0 {
				t.Error("hashCount should be > 0")
			}
		})
	}
}

// Test: Basic round-trip — add items, verify they're found
func TestBloomFilter_AddThenContains(t *testing.T) {
	f := NewBloomFilter(1000, 0.01)

	items := []string{"abc123", "xyz789", "hello", "world", "shrinks"}

	for _, item := range items {
		f.Add(item)
	}

	for _, item := range items {
		if !f.MightContain(item) {
			t.Errorf("MightContain(%q) = false after Add", item)
		}
	}

	if f.itemCount != uint64(len(items)) {
		t.Errorf("itemCount = %d, want %d", f.itemCount, len(items))
	}
}

// Test: No false negatives — the core guarantee
func TestBloomFilter_NoFalseNegatives(t *testing.T) {
	f := NewBloomFilter(10_000, 0.01)

	// Add 1000 items
	for i := 0; i < 1000; i++ {
		f.Add(fmt.Sprintf("item-%d", i))
	}

	// Every single added item MUST return true
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("item-%d", i)
		if !f.MightContain(key) {
			t.Fatalf("False negative for %q — this must NEVER happen", key)
		}
	}
}

// Test: False positive rate is within expected bounds
func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	n := uint64(10_000)
	targetFPR := 0.01
	f := NewBloomFilter(n, targetFPR)

	// Add n items
	for i := uint64(0); i < n; i++ {
		f.Add(fmt.Sprintf("added-%d", i))
	}

	// Check 100,000 items that were NEVER added
	falsePositives := 0
	testCount := 100_000
	for i := 0; i < testCount; i++ {
		// Use a different prefix so these never collide with added items
		if f.MightContain(fmt.Sprintf("never-added-%d", i)) {
			falsePositives++
		}
	}

	observedFPR := float64(falsePositives) / float64(testCount)

	// Allow 2x target as margin for statistical variance
	if observedFPR > targetFPR*2 {
		t.Errorf("False positive rate = %.4f, want < %.4f (2x target of %.4f)",
			observedFPR, targetFPR*2, targetFPR)
	}

	t.Logf("FPR — target: %.4f, observed: %.4f (%d/%d false positives)",
		targetFPR, observedFPR, falsePositives, testCount)
}

// Test: Empty filter returns false for everything
func TestBloomFilter_Empty(t *testing.T) {
	f := NewBloomFilter(1000, 0.01)

	items := []string{"anything", "at", "all", "test123", "abc", "xyz789", "shrinks"}

	for _, item := range items {
		if f.MightContain(item) {
			t.Errorf("Empty filter returned true for %q", item)
		}
	}
}

// Test: Concurrent reads and writes don't corrupt or deadlock
func TestBloomFilter_Concurrency(t *testing.T) {
	f := NewBloomFilter(100_000, 0.01)

	var wg sync.WaitGroup

	// 10 writers, each adding 1000 items
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				f.Add(fmt.Sprintf("writer%d-item%d", id, i))
			}
		}(w)
	}

	// 10 readers, each checking 1000 items
	for r := 0; r < 10; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_ = f.MightContain(fmt.Sprintf("reader%d-query%d", id, i))
			}
		}(r)
	}

	wg.Wait()

	// All 10 writers added 1000 items each
	if f.itemCount != 10_000 {
		t.Errorf("itemCount = %d, want 10000", f.itemCount)
	}

	// Verify no false negatives on the items we added
	for w := 0; w < 10; w++ {
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("writer%d-item%d", w, i)
			if !f.MightContain(key) {
				t.Fatalf("False negative for %q after concurrent writes", key)
			}
		}
	}
}