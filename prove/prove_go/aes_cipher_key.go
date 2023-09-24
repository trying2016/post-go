package post_go

import (
	"encoding/binary"
	"github.com/zeebo/blake3"
)

func toLeBytes(v interface{}) []byte {
	var b8 [8]byte
	switch vv := v.(type) {
	case uint32:
		binary.LittleEndian.PutUint32(b8[:], vv)
		return b8[:4]
	case uint64:
		binary.LittleEndian.PutUint64(b8[:], vv)
		return b8[:]
	default:
		return nil
	}
}

type AesCipher struct {
	Key        []byte
	NonceGroup uint32
	Pow        uint64
}

func NewAesCipherKey(challenge []byte, nonceGroup uint32, pow uint64) []byte {
	hasher := blake3.New()
	hasher.Write(challenge[:])
	hasher.Write(toLeBytes(nonceGroup))
	hasher.Write(toLeBytes(pow))
	key := hasher.Sum(nil)[:16]
	return key
}

func NewLazyAesCipherKey(challenge []byte, nonce, nonceGroup uint32, pow uint64) []byte {
	hasher := blake3.New()
	hasher.Write(challenge[:])
	hasher.Write(toLeBytes(nonceGroup))
	hasher.Write(toLeBytes(pow))
	hasher.Write(toLeBytes(nonce))
	key := hasher.Sum(nil)[:16]
	return key
}
