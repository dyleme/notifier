package service

import (
	"math/rand/v2"
	"strconv"
)

type RandomIntSeq struct{}

const numberOfDigitsToGenerate = 7

func (ri RandomIntSeq) GenereateCode() string {
	return strconv.Itoa(rand.IntN(1 << numberOfDigitsToGenerate)) //nolint:gosec // noneed in security
}
