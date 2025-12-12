package encoding

import (
	"math/big"
)

const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const Base = 62

func Encode(id uint64) string {
	if id == 0 { return string(Alphabet[0]) }

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
