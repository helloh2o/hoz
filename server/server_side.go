package main

import (
	"leango/hoz"
	"flag"
)

var (
	addr     = flag.String("addr", ":10800", "Local hoz listen address")
	password = flag.String("password", "oor-!@adDxS$&(dl/*?", "Cipher password string")
)

func main() {
	s := hoz.NewServer(hoz.Config{
		Addr:   *addr,
		Cipher: *password,
	})
	s.Start()
}
