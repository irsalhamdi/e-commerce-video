package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/big"
	mrand "math/rand"
	"time"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	var b [8]byte
	_, err := crand.Read(b[:])
	if err != nil {
		mrand.Seed(time.Now().UnixNano())
		return
	}
	mrand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func String(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mrand.Intn(len(charset))]
	}
	return string(b)
}

func StringSecure(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		l := big.NewInt(int64(len(charset)))
		num, err := crand.Int(crand.Reader, l)
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}
