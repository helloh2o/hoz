package cipher

import (
	"github.com/xtaci/kcp-go"
	"io"
	"net"
)

func NewAes(key []byte) (Cipher, error) {
	s20 := new(Aes)
	var err error
	l8 := len(key) / 8
	if l8 == 0 {
		panic("key is too short")
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

func (ae *Aes) ReadPackageFrom(from net.Conn, buf []byte, tls bool) ([]byte, error) {
	var data []byte
	var er error
	var n int
	n, er = from.Read(buf)
	if er != nil {
		return nil, er
	}
	/*if !tls {
		data, er = ae.Decrypt(buf[:n])
	} else {
		data = buf[:n]
	}*/
	data, er = ae.Decrypt(buf[:n])
	if er != nil {
		return nil, er
	}
	return data, nil
}

func (ae *Aes) EncryptFromTo(from io.Reader, to io.Writer, tls bool) (n int, err error) {
	defer func() {
		recover()
	}()
	buf := make([]byte, 32*1020)
	for {
		n, er := from.Read(buf)
		if er != nil {
			return n, er
		}
		if n > 0 {
			var data []byte
			var err error
			/*if !tls {
				data, err = ae.Encrypt(buf[:n])
				if err != nil {
					return n, err
				}
			} else {
				data = buf[:n]
			}*/
			data, err = ae.Encrypt(buf[:n])
			if err != nil {
				return n, err
			}
			// write
			n, er = ae.Write(data, to)
			if er != nil {
				return n, er
			}
			//log.Printf("write encrypt data  %d \n", n)
		}
	}
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

func (ae *Aes) Write(data []byte, writer io.Writer) (n int, err error) {
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
