package hoz

type pool struct {
	activeMap map[string]Connection
	idleQueue chan []Connection
}