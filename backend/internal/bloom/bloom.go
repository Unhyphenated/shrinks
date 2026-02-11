package bloom

import (
	"math"
	"sync"
)

type BloomFilter struct {
	bitSet []byte
	bitCount uint64
	hashCount uint64
	itemCount uint64
	mtx sync.RWMutex
}

func NewBloomFilter(expectedItems uint64, falsePositiveRate float64) *BloomFilter {
	n := float64(expectedItems)
	p := falsePositiveRate
	m := math.Ceil(-(n * math.Log(p)) / (math.Ln2 * math.Ln2))

	k := math.Ceil(m / n * math.Ln2)

	return &BloomFilter {
		bitSet: make([]byte, (uint64(m) + 7) / 8),
		bitCount: uint64(m),
		hashCount: uint64(k),
	}
}

func (bf *BloomFilter) Add(shortCode string)