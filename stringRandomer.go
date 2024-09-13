package main

import (
	crypto_rand "crypto/rand"
	"math/big"
)

var (
	ASCII []byte = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

func getRandomShortLink() (string, error) {
	maxLen := 10
	s := make([]byte, maxLen)
	for i := 0; i < maxLen; i++ {
		nbig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(len(ASCII))))
		if err != nil {
			return "", err
		}
		n := int(nbig.Int64())
		s[i] = ASCII[n]
	}
	return string(s), nil

}
