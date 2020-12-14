package cipher

import (
	"encoding/hex"
	"errors"
	"io"
	"net"
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
	xor.trySelfUpdate(len(src))
	for i, b := range src {
		src[i] = b ^ byte(i%255) ^ xor.remainder[i]
	}
	return src, nil
}

func (xor *XORCipher) Decrypt(src []byte) ([]byte, error) {
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

func (xor *XORCipher) ReadPackageFrom(from net.Conn, buf []byte, tls bool) ([]byte, error) {
	var data []byte
	var er error
	var n int
	n, er = from.Read(buf)
	if er != nil {
		return nil, er
	}
	/*if !tls {
		data, _ = xor.Decrypt(buf[:n])
	} else {
		data = buf[:n]
	}*/
	data, _ = xor.Decrypt(buf[:n])
	return data, nil
}

func (xor *XORCipher) EncryptFromTo(from io.Reader, to io.Writer, tls bool) (n int, err error) {
	defer func() {
		recover()
	}()
	buf := make([]byte, 4096)
	for {
		n, er := from.Read(buf)
		if er != nil {
			return n, er
		}
		if n > 0 {
			var data []byte
			/*if !tls {
				data, _ = xor.Encrypt(buf[:n])
			} else {
				data = buf[:n]
			}*/
			data, _ = xor.Encrypt(buf[:n])
			//log.Printf("EncryptFromTo %d \n%v\n", len(endata), endata)
			// write
			n, er = xor.Write(data, to)
			if er != nil {
				return n, er
			}
			//log.Printf("write encrypt data  %d \n", n)
		}
	}
}

func (xor *XORCipher) Write(data []byte, writer io.Writer) (n int, err error) {
	lens := len(data)
	writen := 0
	for {
		n, err = writer.Write(data)
		if err != nil {
			return writen, err
		}
		if n > 0 {
			writen += n
		}
		if writen == lens {
			break
		}
	}
	return lens, nil
}
