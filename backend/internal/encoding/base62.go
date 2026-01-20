package encoding

import (
	"errors"
	"math/big"
)

const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const Base = 62

var decodeMap map[rune]uint64

func init() {
	decodeMap = make(map[rune]uint64)
	for i, char := range Alphabet {
		decodeMap[char] = uint64(i)
	}
}

func Encode(id uint64) string {
	if id == 0 {
		return string(Alphabet[0])
	}

	num := new(big.Int).SetUint64(id)
	var encoded string

	base := big.NewInt(Base)
	remainder := big.NewInt(0)

	for num.Cmp(big.NewInt(0)) > 0 {
		num.DivMod(num, base, remainder)
		charIndex := remainder.Int64()
		encoded = string(Alphabet[charIndex]) + encoded
	}

	return encoded
}

func Decode(encoded string) (uint64, error) {
	if encoded == "" {
		return 0, errors.New("empty string")
	}
	base := big.NewInt(Base)
	result := big.NewInt(0)

	for _, char := range encoded {
		index, exists := decodeMap[char]
		if !exists {
			return 0, errors.New("invalid character")
		}
		idxBig := big.NewInt(int64(index))

		result.Mul(result, base)
		result.Add(result, idxBig)
	}

	return result.Uint64(), nil
}
