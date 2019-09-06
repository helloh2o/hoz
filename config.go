package hoz

import "golang.org/x/time/rate"

type Config struct {
	Addr       string     `json:"addr"`
	RemoteAddr string     `json:"remote_addr"`
	MaxSpeed   rate.Limit `json:"max_speed"`
	MaxTraffic int64      `json:"max_traffic"`
	Cipher  string
}
