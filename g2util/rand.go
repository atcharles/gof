package g2util

import (
	cRand "crypto/rand"
	"math/big"
	"math/rand"
)

// CryptoRandInt ...
func CryptoRandInt(n int) *big.Int {
	b, _ := cRand.Int(cRand.Reader, big.NewInt(int64(n)))
	return b
}

// MathRandInt ...
func MathRandInt(n int) int { return rand.Intn(n) }
