package hoz

import (
	"net"
	"io"
	"bytes"
	"net/http"
	"bufio"
	"strings"
	"time"
	"runtime/debug"
	"fmt"
	"hoz/pkg"
)

type Connection struct {
	reader pkg.PackageReader
	writer pkg.PackageWriter
	conn   net.Conn
	s      *Server
}

func (c *Connection) handle() {
	var remote net.Conn
	var err error
	buf := make([]byte, 32*1024)
	defer func() {
		if r := recover(); r != nil {
			LOG.Printf("Recover from handle, Stack::\n%s\n", debug.Stack())
		}
		c.conn.Close()
		if remote != nil {
			remote.Close()
		}
	}()
	if c.s.RemoteAddr == "" {
		// TODO hoz server side more self handshake
		// read pkg length
		pack, err := c.reader.ReadPackageFrom(c.conn, buf)
		if err != nil {
			LOG.Println(err)
			return
		}
		// parse host
		br := bufio.NewReader(bytes.NewReader(pack))
		req, err := http.ReadRequest(br)
		if err != nil {
			LOG.Printf("ReadRequest error %v\n", err)
			return
		}
		host := req.URL.Host
		if len(host) > 0 && strings.Index(host, ":") == -1 {
			host += ":80"
		} else if host == "" {
			host = fmt.Sprint(req.Header.Get("Shost"), ":", req.Header.Get("Sport"))
		}
		// dial remote
		remote, err = net.DialTimeout("tcp", host, time.Second*3)
		if err != nil {
			LOG.Printf("dial imeout remote error %v\n", err)
			return
		}
		switch req.Method {
		case "SOCKS5":
			// response socks5 established
			ok := c.writeExBytes([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, c.conn)
			if !ok {
				return
			}
		case "CONNECT":
			established := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			established, err = c.s.cipher.Encrypt(established)
			if err != nil {
				return
			}
			_, err = c.writer.Write(established, c.conn)
			if err == nil {
				LOG.Println("Connection established succeed.")
			} else {
				return
			}
		default:
			// write http pack to real host
			_, err = c.writer.Write(pack, remote)
			if err != nil {
				return
			}
			LOG.Println("HTTP write request.")
		}
		// server encrypt remote to client
		go func() {
			c.writer.EncryptFromTo(remote, c.conn)
			c.conn.Close()
		}()
		for {
			// read pkg length
			pack, err = c.reader.ReadPackageFrom(c.conn, buf)
			if err != nil {
				LOG.Println(err)
				return
			}
			_, err = c.writer.Write(pack, remote)
			if err != nil {
				return
			}
		}
	} else {
		// TODO hoz client side more self handshake
		remote, err = net.DialTimeout("tcp", c.s.RemoteAddr, time.Second*5)
		if err != nil {
			LOG.Printf("net dial failed err %s >> %s\n", err.Error(), c.s.RemoteAddr)
			return
		}
		// try handshake socks5
		ok, data, _ := c.handshakeSocks(buf)
		if ok {
			// socks5 read
			ok, data, err = c.parseSocks(buf)
			if ok {
				// send socks5 to http
				ok = c.writeExBytes(data, remote)
				if !ok {
					return
				}
			} else {
				return
			}
		} else if data != nil {
			// http read bytes to remote
			ok = c.writeExBytes(data, remote)
			if !ok {
				return
			}
		} else {
			// socks5 ver check failed
			return
		}
		go func() {
			c.writer.EncryptFromTo(c.conn, remote)
			remote.Close()
		}()
		for {
			pack, err := c.reader.ReadPackageFrom(remote, buf)
			if err != nil {
				//LOG.Printf("Client side closed %s\n", err.Error())
				return
			}
			_, err = c.writer.Write(pack, c.conn)
			if err != nil {
				return
			}
		}
	}
}

func (c *Connection) writeExBytes(data [] byte, remote net.Conn) bool {
	endata, err := c.s.cipher.Encrypt(data)
	if err != nil {
		LOG.Printf("encrypt http data err %v\n", err)
		return false
	}
	_, err = remote.Write(endata)
	if err != nil {
		return false
	}
	return true
}
func (c *Connection) handshakeSocks(buf []byte) (bool, []byte, error) {
	handshake := func(pkg []byte, conn net.Conn) bool {
		ver := pkg[0]
		if ver != 0x05 {
			LOG.Printf("unsupport socks version %d \n", ver)
			return false
		}
		resp := pkg[:0]
		resp = append(resp, 0x05)
		resp = append(resp, 0x00)
		n, err := conn.Write(resp)
		if n != 2 || err != nil {
			return false
		}
		// handshake over
		return true
	}
	n, er := io.ReadAtLeast(c.conn, buf,3)
	if er != nil {
		return false, nil, er
	}
	// socks5
	if buf[0] == 0x05 {
		ok := handshake(buf[:n], c.conn)
		return ok, nil, nil
	}
	// http, buf is left byte
	return false, buf[:n], nil
}

func (c *Connection) parseSocks(buf []byte) (bool, []byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	n, er := io.ReadAtLeast(c.conn, buf, 6)
	c.conn.SetReadDeadline(time.Time{})
	if er != nil {
		return false, nil, er
	}
	// socks5
	if buf[0] == 0x05 {
		to5, ok := parseSocks5Request(buf[:n])
		if !ok {
			c.conn.Write(to5)
			return false, nil, nil
		}
		return true, to5, nil
	}
	// http, buf is left byte
	return false, buf[:n], nil
}
