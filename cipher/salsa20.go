package cipher

import (
	"github.com/xtaci/kcp-go"
	"io"
	"net"
)

func NewSalsa20(key []byte) (Cipher, error) {
	s20 := new(Salsa20)
	var err error
	s20.crypt, err = kcp.NewSalsa20BlockCrypt(key)
	if err != nil {
		return nil, err
	}
	return s20, nil
}

type Salsa20 struct {
	key   []byte
	crypt kcp.BlockCrypt
}

func (s20 *Salsa20) ReadPackageFrom(from net.Conn, buf []byte, tls bool) ([]byte, error) {
	var data []byte
	var er error
	var n int
	n, er = from.Read(buf)
	if er != nil {
		return nil, er
	}
	/*if !tls {
		data, er = s20.Decrypt(buf[:n])
	} else {
		data = buf[:n]
	}*/
	data, er = s20.Decrypt(buf[:n])
	if er != nil {
		return nil, er
	}
	return data, nil
}

func (s20 *Salsa20) EncryptFromTo(from io.Reader, to io.Writer, tls bool) (n int, err error) {
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
				data, err = s20.Encrypt(buf[:n])
				if err != nil {
					return n, err
				}
			} else {
				data = buf[:n]
			}*/
			data, err = s20.Encrypt(buf[:n])
			if err != nil {
				return n, err
			}
			// write
			n, er = s20.Write(data, to)
			if er != nil {
				return n, er
			}
		}
	}
}

func (s20 *Salsa20) Encrypt(src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	s20.crypt.Encrypt(dst, src)
	return dst, nil
}

func (s20 *Salsa20) Decrypt(src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	s20.crypt.Decrypt(dst, src)
	return dst, nil
}

func (s20 *Salsa20) Write(data []byte, writer io.Writer) (n int, err error) {
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
