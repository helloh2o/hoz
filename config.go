package hoz

type Config struct {
	Addr       string `json:"addr"`
	RemoteAddr string `json:"remote_addr"`
	MaxTraffic int64  `json:"max_traffic"`
	Cipher     string
}
