package pkg

import (
	"io"
	"net"
)

type PackageReader interface {
	// tls connection or not
	ReadPackageFrom(from net.Conn, buf []byte, tls bool) ([]byte, error)
}

type PackageWriter interface {
	rawWriter
	// tls connection or not
	EncryptFromTo(from io.Reader, to io.Writer, tls bool) (n int, err error)
}

type rawWriter interface {
	Write(data []byte, writer io.Writer) (n int, err error)
}
