package cipher

import (
	"encoding/binary"
)

type OORR struct {
	SecretKey   []byte
	KeyMaxIndex int
}

func (or *OORR) Encrypt(src []byte) ([]byte, error) {
	length := len(src)
	for i, b := range src {
		if or.KeyMaxIndex >= 0 {
			src[i] = b ^ byte(i) ^ or.SecretKey[i % or.KeyMaxIndex]
		}else {
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
		if or.KeyMaxIndex >= 0 {
			src[i] = b ^ byte(i) ^ or.SecretKey[i % or.KeyMaxIndex]
		}else {
			src[i] = b ^ byte(i)
		}
	}
	return src, nil
}
