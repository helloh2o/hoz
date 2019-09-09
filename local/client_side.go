package main

import (
	"leango/hoz"
	//_ "net/http/pprof"
	//"net/http"
	"flag"
)

var (
	addr     = flag.String("addr", ":1080", "Local hoz listen address")
	remote   = flag.String("remote", "127.0.0.1:10800", "Remote hoz server address")
	password = flag.String("password", "oor-!@adDxS$&(dl/*?", "Cipher password string")
)

func main() {
	s := hoz.NewServer(hoz.Config{
		Addr:       *addr,
		RemoteAddr: *remote,
		Cipher:     *password,
	})
	/*go func() {
		http.ListenAndServe(":6061", nil)
	}()*/
	s.Start()
}
