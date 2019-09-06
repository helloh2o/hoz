package hoz

import (
	"net"
	"strings"
	"leango/hoz/cipher"
)

type Server struct {
	Config
	cipher cipher.Cipher
	ln     net.Listener
}

func NewServer(config Config) *Server {
	return &Server{
		Config: config,
	}
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		LOG.Printf("server startup err %v \n", err)
	}
	LOG.Printf("Server startup, listen on %s\n", s.Config.Addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			LOG.Printf("Accept connection err %v \n", err)
		}
		var nc Connection
		switch {
		case strings.Index(s.Config.Cipher, "oor") == 0:
			key := []byte(s.Config.Cipher[3:])
			s.cipher = &cipher.OORR{SecretKey: key, KeyMaxIndex: len(key) - 1}
		default:
			s.cipher = &cipher.OORR{}
		}
		nc = &Xconn{conn, s}
		go nc.handle()
	}
}
