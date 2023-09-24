package post

// #include "post.h"
import "C"

type Aes struct {
	aes *C.Aes
}

func NewAes(key []byte) *Aes {
	return &Aes{
		aes: CreateAes(key),
	}
}

func (a *Aes) Encrypt(input []byte, output []byte, batchSize int) {
	EncryptAes(a.aes, input, output, batchSize)
}

func (a *Aes) EncryptUint(input []byte, output []uint64, batchSize int) {
	EncryptAesUint(a.aes, input, output, batchSize)
}

func (a *Aes) Free() {
	FreeAes(a.aes)
}
