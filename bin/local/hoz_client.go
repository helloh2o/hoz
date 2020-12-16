package main

import (
	//_ "net/http/pprof"
	//"net/http"
	"flag"
	"hoz"
)

var (
	addr     = flag.String("addr", ":1080", "Local hoz listen address")
	kcp      = flag.Bool("kcp", false, "use kcp protocol")
	remote   = flag.String("remote", "127.0.0.1:10800", "Remote hoz server address")
	password = flag.String("password", "little://!@adDxS$&(dl/*?QKc$mJ?PdTkajGzSNMILH{t4_hvFR>", "Cipher password string")
)

func main() {
	flag.Parse()
	s := hoz.NewServer(hoz.Config{
		Addr:       *addr,
		RemoteAddr: *remote,
		Cipher:     *password,
		KCP:        *kcp,
	})
	/*go func() {
		http.ListenAndServe(":6061", nil)
	}()*/
	s.Start()
}
