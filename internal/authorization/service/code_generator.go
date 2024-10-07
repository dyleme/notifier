package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type RandomIntSeq struct {
	MaxNumber *big.Int
}

const maxNumber = 1_000_000

func NewRandomIntSeq() RandomIntSeq {
	return RandomIntSeq{MaxNumber: big.NewInt(maxNumber)}
}

func (ri RandomIntSeq) GenereateCode() string {
	bigInt, err := rand.Int(rand.Reader, ri.MaxNumber)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%06d", bigInt.Int64())
}
