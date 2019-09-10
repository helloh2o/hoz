package rwder

import (
	"encoding/binary"
	"hoz/cipher"
	"net"
	"io"
	"hoz/pkg"
)

type OorReader struct {
	cipher cipher.Cipher
}

func NewOorReader(cipher cipher.Cipher) pkg.PackageReader {
	return &OorReader{cipher: cipher}
}

func NewOorWriter(cipher cipher.Cipher) pkg.PackageWriter {
	return &OorWriter{cipher: cipher}
}

type OorWriter struct {
	cipher cipher.Cipher
}


func (r *OorReader) ReadPackageFrom(from net.Conn, buf []byte) ([]byte, error) {
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
	data, er = r.cipher.Decrypt(data)
	//log.Printf("raw data is \n %s\n", string(data))
	if er != nil {
		return nil, er
	}
	return data, nil
}


func (w *OorWriter) EncryptFromTo(from io.Reader, to io.Writer) (n int, err error) {
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
			endata, err := w.cipher.Encrypt(buf[:n])
			if err != nil {
				return n, err
			}
			//log.Printf("EncryptFromTo %d \n%v\n", len(endata), endata)
			// write
			n, er = w.Write(endata, to)
			if er != nil {
				return n, er
			}
			//log.Printf("write encrypt data  %d \n", n)
		}
	}
}

func (w *OorWriter) Write(data []byte, writer io.Writer) (n int, err error) {
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