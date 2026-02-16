package bloom

import (
	"hash/fnv"
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
	if expectedItems <= 0 {
		expectedItems = 1000 // Avoid division by zero
	}

	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01
	}

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

func (bf *BloomFilter) Add(shortCode string) {
	h1, h2 := bf.hashes(shortCode)

	bf.mtx.Lock()
	defer bf.mtx.Unlock()
	for i := uint64(0); i < bf.hashCount; i++ {
		pos := (h1 + i * h2) % bf.bitCount
		bf.bitSet[pos / 8] |= (1 << (pos % 8))
	}
	bf.itemCount++
}

func (bf *BloomFilter) MightContain(shortCode string) bool {
	h1, h2 := bf.hashes(shortCode)

	bf.mtx.RLock()
	defer bf.mtx.RUnlock()
	for i := uint64(0); i < bf.hashCount; i++ {
		pos := (h1 + i * h2) % bf.bitCount
		if (bf.bitSet[pos / 8] & (1 << (pos % 8)) == 0) {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) hashes(shortCode string) (uint64, uint64) {
	h1, h2 := fnv.New64(), fnv.New64a()
	h1.Write([]byte(shortCode))
	h2.Write([]byte(shortCode))

	return h1.Sum64(), h2.Sum64()
}