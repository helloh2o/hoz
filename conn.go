package hoz

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/xtaci/kcp-go"
	"hoz/cipher"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Connection struct {
	conn net.Conn
	s    *Server
}

func (c *Connection) handle() {
	defer func() {
		if r := recover(); r != nil {
			LOG.Printf("Recover from handle, %v, Stack::\n%s\n", r, debug.Stack())
		}
		_ = c.conn.Close()
	}()
	if c.s.RemoteAddr == "" {
		c.serverSide()
	} else {
		c.clientSide()
	}
}

func (server *Connection) serverSide() {
	var remote net.Conn
	var err error
	buf := make([]byte, 4096)
	n, err := server.conn.Read(buf)
	if err != nil {
		LOG.Println("serverSide read first time error ", err)
		return
	}
	// decode
	data, _ := server.s.cipher.Decrypt(buf[:n])
	// parse host
	br := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(br)
	if err != nil {
		LOG.Printf("Http ReadRequest error %v\n", err)
		return
	}
	host := req.URL.Host
	if len(host) > 0 && strings.Index(host, ":") == -1 {
		host += ":80"
	} else if host == "" {
		host = fmt.Sprint(req.Header.Get("Shost"), ":", req.Header.Get("Sport"))
	}
	LOG.Println("try connect real host::" + host)
	// dial remote
	remote, err = net.DialTimeout("tcp", host, time.Second*5)
	if err != nil {
		LOG.Printf("dial imeout real remote error %v\n", err)
		return
	}
	defer func() {
		_ = remote.Close()
		_ = server.conn.Close()
	}()
	var established []byte
	switch req.Method {
	case "SOCKS5":
		// response socks5 established
		established = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case "CONNECT":
		established = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
	default:
		// write http pack to real host
		_, err = remote.Write(data)
		if err != nil {
			LOG.Println("Write HTTP header to remote error")
			return
		}
		LOG.Println("HTTP write request.")
	}
	if len(established) > 0 {
		established, _ := server.s.cipher.Encrypt(established)
		_, err = server.conn.Write(established)
		if err != nil {
			LOG.Println("write established error ", err)
			return
		}
	}
	pipe(server.conn, remote, server.s.cipher, false)
}

func (client *Connection) clientSide() {
	var remote net.Conn
	var err error
	if client.s.Config.KCP {
		remote, err = kcp.DialWithOptions(client.s.RemoteAddr, nil, 10, 3)
	} else {
		remote, err = net.DialTimeout("tcp", client.s.RemoteAddr, time.Second*10)
	}
	if err != nil {
		LOG.Printf("net dial failed err %s >> %s\n", err.Error(), client.s.RemoteAddr)
		return
	}
	defer func() {
		_ = remote.Close()
		_ = client.conn.Close()
	}()
	// try handshake socks5
	buf := make([]byte, 81920)
	ok, data, _ := client.handshakeSocks(buf)
	if ok {
		// socks5 read
		ok, data, err = client.parseSocks(buf)
		if ok {
			// send socks5 to http
			ok = client.writeExBytes(data, remote)
			if !ok {
				return
			}
		} else {
			return
		}
	} else if data != nil {
		//LOG.Println(string(data))
		// http read bytes to remote
		ok = client.writeExBytes(data, remote)
		if !ok {
			return
		}
	} else {
		// socks5 ver check failed
		return
	}
	pipe(client.conn, remote, client.s.cipher, true)
}

func pipe(local, remote net.Conn, cp cipher.Cipher, localSide bool) {
	defer func() {
		_ = local.Close()
		_ = remote.Close()
	}()
	var errChan = make(chan error)
	go func() {
		buf1 := make([]byte, 81920)
		for {
			// copy remote <=> local <=> client
			n, err := remote.Read(buf1)
			if err != nil {
				LOG.Println("remote read error ", err)
				errChan <- err
				break
			}
			// decode
			var pack []byte
			if localSide {
				pack, _ = cp.Decrypt(buf1[:n])
			} else {
				pack, _ = cp.Encrypt(buf1[:n])
			}
			_, err = local.Write(pack)
			if err != nil {
				LOG.Println("copy remote to client error ", err)
				errChan <- err
			}
		}

	}()
	go func() {
		buf2 := make([]byte, 4096)
		for {
			n, err := local.Read(buf2)
			if err != nil {
				LOG.Println("local read error ", err)
				LOG.Println("local remote addr  ", local.RemoteAddr())
				errChan <- err
				break
			}
			var pack []byte
			// encode to remote
			if localSide {
				pack, _ = cp.Encrypt(buf2[:n])
			} else {
				pack, _ = cp.Decrypt(buf2[:n])
			}
			n, err = remote.Write(pack)
			if err != nil {
				LOG.Println("copy client to remote error ", err)
				errChan <- err
				break
			}
		}
	}()
	err := <-errChan
	LOG.Println("pipe end err::", err)
}

func (c *Connection) writeExBytes(data []byte, remote net.Conn) bool {
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
