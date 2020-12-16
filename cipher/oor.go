package cipher

import (
	"encoding/hex"
	"errors"
	"sync"
)

type XORCipher struct {
	sync.RWMutex
	password    []byte
	pwdMaxIndex int
	remainder   []byte
	maxLen      int
}

// 简单的混淆加密 maxBufferLen 加密数据最大长度
func NewXORCipher(key string) (*XORCipher, error) {
	if len(key) < 4 {
		return nil, errors.New("XOR key must more than 4 characters")
	}
	or := &XORCipher{}
	or.password = []byte(hex.EncodeToString([]byte(key)))
	or.pwdMaxIndex = len(or.password) - 1
	or.updateMaxLen(8192)
	return or, nil
}

func (xor *XORCipher) updateMaxLen(max int) {
	xor.Lock()
	defer xor.Unlock()
	xor.maxLen = max
	for i := 0; i < xor.maxLen; i++ {
		xor.remainder = append(xor.remainder, xor.password[i%xor.pwdMaxIndex])
	}
}

func (xor *XORCipher) Encrypt(src []byte) ([]byte, error) {
	//return src, nil
	xor.trySelfUpdate(len(src))
	for i, b := range src {
		src[i] = b ^ byte(i%255) ^ xor.remainder[i]
	}
	return src, nil
}

func (xor *XORCipher) Decrypt(src []byte) ([]byte, error) {
	//return src, nil
	xor.trySelfUpdate(len(src))
	for i, b := range src {
		src[i] = b ^ byte(i%255) ^ xor.remainder[i]
	}
	return src, nil
}

// 更新长度
func (xor *XORCipher) trySelfUpdate(length int) {
	if length < 8192 {
		return
	}
	xor.RLock()
	if xor.maxLen < length {
		xor.RUnlock()
		xor.updateMaxLen(length)
	} else {
		xor.RUnlock()
	}
}
