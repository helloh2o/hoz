package cipher

import (
	"github.com/xtaci/kcp-go"
)

func NewAes(key []byte) (Cipher, error) {
	s20 := new(Aes)
	var err error
	l8 := len(key) / 8
	if l8 < 2 {
		panic("key is too short")
	}
	if l8 > 4 {
		l8 = 3
	}
	key = key[:8*l8]
	s20.crypt, err = kcp.NewAESBlockCrypt(key)
	if err != nil {
		return nil, err
	}
	return s20, nil
}

type Aes struct {
	key   []byte
	crypt kcp.BlockCrypt
}

func (ae *Aes) Encrypt(src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	ae.crypt.Encrypt(dst, src)
	return dst, nil
}

func (ae *Aes) Decrypt(src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	ae.crypt.Decrypt(dst, src)
	return dst, nil
}
