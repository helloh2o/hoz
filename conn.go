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
		// TODO handshake
		// parse host
		br := bufio.NewReader(bytes.NewReader(data))
		req, err := http.ReadRequest(br)
		if err != nil {
			LOG.Printf("ReadRequest error %v\n", err)
			return
		}
		host := req.URL.Host
		if strings.Index(host, ":") == -1 {
			host += ":80"
		}
		if req.Method == "CONNECT" {
			established := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			established, err = c.s.cipher.Encrypt(established)
			if err != nil {
				return
			}
			_, err = c.pkgWriteTo(established, c.conn)
			if err == nil {
				LOG.Println("Notify Connection established succeed.")
			}
		}
		// dial remote
		remote, err = net.DialTimeout("tcp", host, time.Second*5)
		if err != nil {
			LOG.Printf("DialTimeout remote error %v\n", err)
			return
		}
		if req.Method != "CONNECT" {
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
		// TODO handshake
		//Encrypt Client Side
		remote, err := net.DialTimeout("tcp", c.s.RemoteAddr, time.Second*5)
		if err != nil {
			LOG.Printf("net dial failed err %s >> %s\n", err.Error(), c.s.RemoteAddr)
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
