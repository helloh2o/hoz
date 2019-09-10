package cipher

import (
	"log"
	"testing"
	"time"
)

// 1GB data test
func TestOORR_Encrypt(t *testing.T) {
	data := []byte("12")
	o := NewOor([]byte{})
	b := time.Now().Unix()
	total := 1048576 * 512 * 2
	for i := 0; i < 1048576*512; i++ {
		o.Encrypt(data)
	}
	e := time.Now().Unix()
	// two byte small package 35M/s
	log.Printf("encrypte seconds %d, Total bytes %d \n", e-b, total)
}
