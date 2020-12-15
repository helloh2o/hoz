package main

import (
	"flag"
	"hoz"
)

var (
	addr     = flag.String("addr", ":10800", "Local hoz listen address")
	password = flag.String("password", "oor://!@adDxS$&(dl/*?QKc$mJ?PdTkajGzSNMILH{t4_hvFR>", "Cipher password string")
)

func main() {
	flag.Parse()
	s := hoz.NewServer(hoz.Config{
		Addr:   *addr,
		Cipher: *password,
		KCP:    true,
	})
	s.Start()
}
