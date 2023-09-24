package post

import (
	"crypto/aes"
	"crypto/sha256"
	"testing"
)

func TestAes128(t *testing.T) {
	key := sha256.Sum256([]byte(""))
	block, err := aes.NewCipher(key[:16])
	if err != nil {
		t.Fatal(err)
	}
	var b1 [32]byte
	var b2 [32]byte
	block.Encrypt(b1[:], b2[:])
}
