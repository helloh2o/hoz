package hoz

import (
	"net"
	"io"
	"encoding/binary"
	"bytes"
	"net/http"
	"bufio"
	"strings"
	"time"
	"runtime/debug"
	"fmt"
)

type Connection interface {
	handle()
}

type Xconn struct {
	conn net.Conn
	s    *Server
}

func (c *Xconn) handle() {
	var remote net.Conn
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
	//Decrypt Server side
	if c.s.RemoteAddr == "" {
		// read pkg length
		data, err := c.readPackageFrom(c.conn, buf)
		if err != nil {
			LOG.Println(err)
			return
		}
		// TODO more self handshake
		// parse host
		br := bufio.NewReader(bytes.NewReader(data))
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
		if req.Method == "CONNECT" {
			established := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			established, err = c.s.cipher.Encrypt(established)
			if err != nil {
				return
			}
			_, err = c.pkgWriteTo(established, c.conn)
			if err == nil {
				LOG.Println("Connection established succeed.")
			} else {
				return
			}
		}
		// dial remote
		remote, err = net.DialTimeout("tcp", host, time.Second*5)
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
			// do nothing
		default:
			// write pkg to real host
			_, err = c.pkgWriteTo(data, remote)
			if err != nil {
				return
			}
		}
		// server encrypt remote to client
		go func() {
			c.encryptFromTo(remote, c.conn)
			c.conn.Close()
		}()
		for {
			// read pkg length
			data, err = c.readPackageFrom(c.conn, buf)
			if err != nil {
				LOG.Println(err)
				return
			}
			_, err = c.pkgWriteTo(data, remote)
			if err != nil {
				return
			}
		}
	} else {
		// TODO more self handshake
		//Encrypt Client Side
		remote, err := net.DialTimeout("tcp", c.s.RemoteAddr, time.Second*5)
		if err != nil {
			LOG.Printf("net dial failed err %s >> %s\n", err.Error(), c.s.RemoteAddr)
			return
		}
		// try handshake socks5
		ok, data, err := c.handshakeSocks()
		if ok {
			// socks5 read
			ok, data, err = c.parseSocks()
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
			c.encryptFromTo(c.conn, remote)
			remote.Close()
		}()
		for {
			pkg, err := c.readPackageFrom(remote, buf)
			if err != nil {
				//LOG.Printf("Client side closed %s\n", err.Error())
				return
			}
			_, err = c.pkgWriteTo(pkg, c.conn)
			if err != nil {
				return
			}
		}
	}
}

func (c *Xconn) writeExBytes(data [] byte, remote net.Conn) bool {
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
func (c *Xconn) handshakeSocks() (bool, []byte, error) {
	buf := make([]byte, 1024)
	n, er := io.ReadAtLeast(c.conn, buf, 3)
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

func (c *Xconn) parseSocks() (bool, []byte, error) {
	buf := make([]byte, 128)
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

func (c *Xconn) readPackageFrom(from net.Conn, buf []byte) ([]byte, error) {
	n, er := io.ReadFull(from, buf[:4])
	if er != nil {
		return nil, er
	}
	pkgLen := binary.BigEndian.Uint32(buf[:4])
	//LOG.Printf("Read Package Len %d\n", pkgLen)
	n, er = io.ReadFull(from, buf[:pkgLen])
	if er != nil {
		//LOG.Printf("Has read size %d\n", n)
		return nil, er
	}
	data := buf[:n]
	data, er = c.s.cipher.Decrypt(data)
	if er != nil {
		return nil, er
	}
	return data, nil
}
func (c *Xconn) encryptFromTo(from io.Reader, to io.Writer) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			LOG.Printf("recover from encryptFromTo, %v \n", debug.Stack())
		}
	}()
	buf := make([]byte, 32*1024-4)
	for {
		n, er := from.Read(buf)
		if er != nil {
			return n, er
		}
		if n > 0 {
			endata, err := c.s.cipher.Encrypt(buf[:n])
			if err != nil {
				return n, err
			}
			//LOG.Printf("encryptFromTo %d \n%v\n", len(endata), endata)
			n, er = c.pkgWriteTo(endata, to)
			if er != nil {
				return n, er
			}
			//LOG.Printf("write encrypt data  %d \n", n)

		}
	}
}
func (c *Xconn) pkgWriteTo(data []byte, writer io.Writer) (n int, err error) {
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
	//LOG.Println("pkgWriteTo ok")
	return lens, nil
}
