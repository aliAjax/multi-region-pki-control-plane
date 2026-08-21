package domain

import (
	"encoding/hex"
	"errors"
	"math/big"
)

type SerialNumber struct{ n *big.Int }

func NewSerial(raw []byte) (SerialNumber, error) {
	if len(raw) < 16 || len(raw) > 20 {
		return SerialNumber{}, errors.New("serial must be 128-160 bits")
	}
	raw = append([]byte(nil), raw...)
	raw[0] &= 0x7f
	if allZero(raw) {
		raw[len(raw)-1] = 1
	}
	return SerialNumber{n: new(big.Int).SetBytes(raw)}, nil
}
func ParseSerial(s string) (SerialNumber, error) {
	b, e := hex.DecodeString(s)
	if e != nil {
		return SerialNumber{}, e
	}
	return NewSerial(b)
}
func (s SerialNumber) String() string {
	if s.n == nil {
		return ""
	}
	return hex.EncodeToString(s.n.Bytes())
}
func (s SerialNumber) BigInt() *big.Int {
	if s.n == nil {
		return nil
	}
	return new(big.Int).Set(s.n)
}
func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
