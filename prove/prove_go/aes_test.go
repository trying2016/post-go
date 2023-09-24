package post_go

import (
	"crypto/aes"
	"testing"
)

func Aes128Encrypt(key, in []byte) ([]byte, error) {
	hash, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(in))
	hash.Encrypt(out, in)
	return out, nil
}

func TestAes128(t *testing.T) {
	// 16 bytes
	key := []byte{116, 14, 225, 154, 198, 165, 75, 70, 183, 102, 247, 44, 171, 175, 107, 76}
	in := []byte{68, 147, 113, 158, 232, 185, 224, 236, 161, 131, 101, 213, 9, 224, 99, 170, 53, 36, 197, 9, 130, 188, 55, 97, 78, 171, 68, 30, 181, 45, 110, 246, 47, 51, 224, 151, 64, 182, 43, 116, 134, 101, 205, 62, 133, 116, 148, 233, 47, 172, 247, 252, 114, 136, 1, 87, 119, 72, 226, 64, 48, 145, 147, 137}
	out := []byte{0, 0, 0, 0, 0, 0, 0, 0, 68, 147, 113, 158, 232, 185, 224, 236, 161, 131, 101, 213, 9, 224, 99, 170, 53, 36, 197, 9, 130, 188, 55, 97, 78, 171, 68, 30, 181, 45, 110, 246, 47, 51, 224, 151, 64, 182, 43, 116, 134, 101, 205, 62, 133, 116, 148, 233, 47, 172, 247, 252, 114, 136, 1, 87, 119, 72, 226, 64, 48, 145, 147, 137}
	encrypt, err := Aes128Encrypt(key, in[:16])
	if err != nil {
		t.Fatal(err)
	}
	if len(encrypt) != len(out) {
		t.Fatal("encrypt length not equal out length")
	}
}
