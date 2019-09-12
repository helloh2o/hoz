package cipher

import (
	"encoding/binary"
	"github.com/xtaci/kcp-go"
	"net"
	"io"
)

func NewAes(key []byte) (Cipher, error) {
	s20 := new(Aes)
	var err error
	if len(key) > 16 {
		key = key[:16]
	}
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
	if !tls {
		n, er = io.ReadFull(from, buf[:4])
		if er != nil {
			return nil, er
		}
		pkgLen := binary.BigEndian.Uint32(buf[:4])
		n, er = io.ReadFull(from, buf[:pkgLen])
		if er != nil {
			return nil, er
		}
		data, er = ae.Decrypt(buf[:n])
	} else {
		n, er = from.Read(buf)
		if er != nil {
			return nil, er
		}
		data = buf[:n]
	}
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
			if !tls {
				data, err = ae.Encrypt(buf[:n])
				if err != nil {
					return n, err
				}
			} else {
				data = buf[:n]
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
	ae.crypt.Encrypt(src, src)
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(len(src)))
	data := append(head, src...)
	return data, nil
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
