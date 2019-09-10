package pkg

import (
	"io"
	"net"
)

type PackageReader interface {
	ReadPackageFrom(from net.Conn, buf []byte) ([]byte, error)
}

type PackageWriter interface {
	rawWriter
	EncryptFromTo(from io.Reader, to io.Writer) (n int, err error)
}

type rawWriter interface {
	Write(data []byte, writer io.Writer) (n int, err error)
}
