package cipher

import (
	"encoding/binary"
	"net"
	"io"
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
	if or.index > 0 {
		for i := 0; i < 32*1024; i++ {
			or.remainder = append(or.remainder, or.password[i%or.index])
		}
	}
	return or
}

func (or *OORR) Encrypt(src []byte) ([]byte, error) {
	length := len(src)
	for i, b := range src {
		if or.index > 0 {
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
		if or.index > 0 {
			src[i] = b ^ byte(i) ^ or.remainder[i]
		} else {
			src[i] = b ^ byte(i)
		}
	}
	return src, nil
}

func (or *OORR) ReadPackageFrom(from net.Conn, buf []byte) ([]byte, error) {
	n, er := io.ReadFull(from, buf[:4])
	if er != nil {
		return nil, er
	}
	pkgLen := binary.BigEndian.Uint32(buf[:4])
	//log.Printf("Read Package Len %d\n", pkgLen)
	n, er = io.ReadFull(from, buf[:pkgLen])
	if er != nil {
		//log.Printf("Has read size %d\n", n)
		return nil, er
	}
	data := buf[:n]
	data, er = or.Decrypt(data)
	//log.Printf("raw data is \n %s\n", string(data))
	if er != nil {
		return nil, er
	}
	return data, nil
}

func (or *OORR) EncryptFromTo(from io.Reader, to io.Writer) (n int, err error) {
	defer func() {
		recover()
	}()
	buf := make([]byte, 32*1024-4)
	for {
		n, er := from.Read(buf)
		if er != nil {
			return n, er
		}
		if n > 0 {
			endata, err := or.Encrypt(buf[:n])
			if err != nil {
				return n, err
			}
			//log.Printf("EncryptFromTo %d \n%v\n", len(endata), endata)
			// write
			n, er = or.Write(endata, to)
			if er != nil {
				return n, er
			}
			//log.Printf("write encrypt data  %d \n", n)
		}
	}
}

func (or *OORR) Write(data []byte, writer io.Writer) (n int, err error) {
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
