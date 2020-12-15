package hoz

import (
	"errors"
	"hoz/cipher"
	"net"
	"strings"
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
	pass := strings.Split(s.Config.Cipher, "://")
	if len(pass) != 2 {
		LOG.Fatal(errors.New("Cipher must be like scheme://password "))
		return
	}
	var waper cipher.Cipher
	var cipherName = pass[0]
	var key = pass[1]
	switch cipherName {
	case "oor":
		waper, err = cipher.NewXORCipher(key)
	case "aes":
		waper, err = cipher.NewAes([]byte(key))
		if err != nil {
			LOG.Fatalf("Init aes cipher error %v\n", err)
		}
	default:
		LOG.Fatalf("Unsuport cipher %s \n", cipherName)
	}
	LOG.Printf("cipher_name=%s, password=%s\n", cipherName, key)
	s.cipher = waper
	LOG.Printf("Server startup, listen on %s\n", s.Config.Addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			LOG.Printf("Accept connection err %v \n", err)
			panic(err)
		}
		nc := &Connection{s: s, conn: conn}
		go nc.handle()
	}
}
