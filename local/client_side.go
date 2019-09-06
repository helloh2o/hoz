package main

import (
	"github.com/google/gops/agent"
	"hoz"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}
	s := hoz.NewServer(hoz.Config{
		Addr: ":1080",
		//RemoteAddr: "127.0.0.1:10800",
		RemoteAddr: "193.110.203.47:10800",
		Cipher:     "oor-!@adDxS$&(dl/*?",
	})
	go func() {
		http.ListenAndServe(":6061", nil)
	}()
	s.Start()
}
