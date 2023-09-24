package codec

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type Post struct {
	Nonce   uint32
	Indices []byte
	Pow     uint64
}

// EncodeScale implements scale codec interface.
func (p *Post) EncodeScale(enc *Encoder) (total int, err error) {
	{
		n, err := EncodeCompact32(enc, uint32(p.Nonce))
		if err != nil {
			return total, err
		}
		total += n
	}
	{
		n, err := EncodeByteSliceWithLimit(enc, p.Indices, 8000) // needs to hold K2*8 bytes at most
		if err != nil {
			return total, err
		}
		total += n
	}
	{
		n, err := EncodeCompact64(enc, p.Pow)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// DecodeScale implements scale codec interface.
func (p *Post) DecodeScale(dec *Decoder) (total int, err error) {
	{
		field, n, err := DecodeCompact32(dec)
		if err != nil {
			return total, err
		}
		total += n
		p.Nonce = field
	}
	{
		field, n, err := DecodeByteSliceWithLimit(dec, 8000) // needs to hold K2*8 bytes at most
		if err != nil {
			return total, err
		}
		total += n
		p.Indices = field
	}
	{
		field, n, err := DecodeCompact64(dec)
		if err != nil {
			return total, err
		}
		total += n
		p.Pow = field
	}
	return total, nil
}

func TestCodecEncode(t *testing.T) {
	var poof = &Post{
		Nonce:   0xffff,
		Indices: []byte("123"),
		Pow:     0xffffff,
	}
	data, err := Encode(poof)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(data) == "feff03000c313233feffff03")
}
func TestCodecDecode(t *testing.T) {
	var post = &Post{}
	data, err := hex.DecodeString("feff03000c313233feffff03")
	if err != nil {
		t.Fatal()
	}
	err = Decode(data, post)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(post.Nonce == 0xffff &&
		post.Pow == 0xffffff && bytes.Equal(post.Indices, []byte("123")))
}
