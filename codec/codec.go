package codec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
)

var ErrShortRead = errors.New("decode from buffer: not all bytes were consumed")

// TODO(dshulyak) this is a temporary solution to improve encoder allocations.
// if this will stay it must be changed to one of the:
// - use buffer with allocations that can be adjusted using stats
// - use multiple buffers that increase in size (e.g. 16, 32, 64, 128 bytes).
var encoderPool = sync.Pool{
	New: func() interface{} {
		b := new(bytes.Buffer)
		b.Grow(64)
		return b
	},
}

func getEncoderBuffer() *bytes.Buffer {
	return encoderPool.Get().(*bytes.Buffer)
}

func putEncoderBuffer(b *bytes.Buffer) {
	b.Reset()
	encoderPool.Put(b)
}

// EncodeTo encodes value to a writer stream.
func EncodeTo(w io.Writer, value Encodable) (int, error) {
	return value.EncodeScale(NewEncoder(w, WithEncodeMaxNested(6)))
}

// Encode value to a byte buffer.
func Encode(value Encodable) ([]byte, error) {
	b := getEncoderBuffer()
	defer putEncoderBuffer(b)
	n, err := EncodeTo(b, value)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, n)
	copy(buf, b.Bytes())
	return buf, nil
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader, opts ...decoderOpts) *Decoder {
	d := &Decoder{
		r:           r,
		maxNested:   MaxNested,
		maxElements: MaxElements,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// DecodeFrom decodes a value using data from a reader stream.
func DecodeFrom(r io.Reader, value Decodable) (int, error) {
	return value.DecodeScale(NewDecoder(r, WithDecodeMaxNested(6)))
}

// Decode value from a byte buffer.
func Decode(buf []byte, value Decodable) error {
	n, err := DecodeFrom(bytes.NewBuffer(buf), value)
	if err != nil {
		return fmt.Errorf("decode from buffer: %w", err)
	}
	if n != len(buf) {
		return ErrShortRead
	}
	return nil
}
