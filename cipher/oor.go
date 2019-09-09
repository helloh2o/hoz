package cipher

import (
	"encoding/binary"
)

type OORR struct {
	password  []byte
	index     int
	remainder []byte
}

func NewOor(key []byte) *OORR {
	or := new(OORR)
	or.password = key
	or.index = len(key) - 1
	if or.index >= 0 {
		for i := 0; i < 32*1024; i++ {
			or.remainder = append(or.remainder, or.password[i%or.index])
		}
	}
	return or
}

func (or *OORR) Encrypt(src []byte) ([]byte, error) {
	length := len(src)
	for i, b := range src {
		if or.index >= 0 {
			src[i] = b ^ byte(i) ^ or.remainder[i]
		} else {
			src[i] = b ^ byte(i)
		}
	}
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(length))
	data := src[:0]
	data = append(head, src...)
	return data, nil
}

func (or *OORR) Decrypt(src []byte) ([]byte, error) {
	for i, b := range src {
		if or.index >= 0 {
			src[i] = b ^ byte(i) ^ or.remainder[i]
		} else {
			src[i] = b ^ byte(i)
		}
	}
	return src, nil
}
