package main

import (
	"leango/hoz"
	"github.com/google/gops/agent"
	"log"
	_ "net/http/pprof"
	"net/http"
)

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}
	s := hoz.NewServer(hoz.Config{
		Addr:   ":10800",
		Cipher: "oor-!@adDxS$&(dl/*?",
	})
	go func() {
		http.ListenAndServe(":6062", nil)
	}()
	s.Start()
}
