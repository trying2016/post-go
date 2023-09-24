package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"reflect"
)

const (
	// MaxElements is the maximum number of elements allowed in a collection if not set explicitly during encoding/decoding.
	MaxElements uint32 = 1 << 20

	// MaxNested is the maximum nested level allowed if not set explicitly during encoding/decoding.
	MaxNested uint = 4
)

var (
	// ErrEncodeTooManyElements is returned when scale limit tag is used and collection has too many elements to encode.
	ErrEncodeTooManyElements = errors.New("too many elements to encode in collection with scale limit set")

	// ErrEncodeNestedTooDeep is returned when the depth of nested types exceeds the limit.
	ErrEncodeNestedTooDeep = errors.New("nested level is too deep")
)

type Encodable interface {
	EncodeScale(*Encoder) (int, error)
}

type encoderOpts func(*Encoder)

// WithEncodeMaxNested sets the nested level of the encoder.
// A value of 0 means no nesting is allowed. The default value is 4.
func WithEncodeMaxNested(nested uint) encoderOpts {
	return func(e *Encoder) {
		e.maxNested = nested
	}
}

// WithEncodeMaxElements sets the maximum number of elements allowed in a collection.
// The default value is 1 << 20.
func WithEncodeMaxElements(elements uint32) encoderOpts {
	return func(e *Encoder) {
		e.maxElements = elements
	}
}

// NewEncoder returns a new encoder that writes to w.
// If w implements io.StringWriter, the returned encoder will be more efficient in encoding strings.
func NewEncoder(w io.Writer, opts ...encoderOpts) *Encoder {
	e := &Encoder{
		w:           w,
		maxNested:   MaxNested,
		maxElements: MaxElements,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

type Encoder struct {
	w           io.Writer
	scratch     [9]byte
	maxNested   uint
	maxElements uint32
}

func (e *Encoder) enterNested() error {
	if e.maxNested == 0 {
		return ErrEncodeNestedTooDeep
	}
	e.maxNested--
	return nil
}

func (e *Encoder) leaveNested() {
	e.maxNested++
}

func EncodeByteSlice(e *Encoder, value []byte) (int, error) {
	return EncodeByteSliceWithLimit(e, value, e.maxElements)
}

func EncodeByteSliceWithLimit(e *Encoder, value []byte, limit uint32) (int, error) {
	total, err := EncodeLen(e, uint32(len(value)), limit)
	if err != nil {
		return 0, err
	}
	n, err := EncodeByteArray(e, value)
	if err != nil {
		return 0, err
	}
	return total + n, nil
}

func EncodeByteArray(e *Encoder, value []byte) (int, error) {
	return e.w.Write(value)
}

func EncodeString(e *Encoder, value string) (int, error) {
	return EncodeStringWithLimit(e, value, e.maxElements)
}

func EncodeStringWithLimit(e *Encoder, value string, limit uint32) (int, error) {
	total, err := EncodeLen(e, uint32(len(value)), limit)
	if err != nil {
		return 0, err
	}
	n, err := io.WriteString(e.w, value)
	if err != nil {
		return 0, err
	}
	return total + n, nil
}

func EncodeStringSlice(e *Encoder, value []string) (int, error) {
	return EncodeStringSliceWithLimit(e, value, e.maxElements)
}

func EncodeStringSliceWithLimit(e *Encoder, value []string, limit uint32) (int, error) {
	valueToBytes := make([][]byte, 0, len(value))
	for i := range value {
		valueToBytes = append(valueToBytes, stringToBytes(value[i]))
	}
	total, err := EncodeLen(e, uint32(len(valueToBytes)), limit)
	if err != nil {
		return 0, fmt.Errorf("EncodeLen failed: %w", err)
	}
	for _, byteSlice := range valueToBytes {
		n, err := EncodeByteSliceWithLimit(e, byteSlice, e.maxElements)
		if err != nil {
			return 0, fmt.Errorf("EncodeByteSliceWithLimit failed: %w", err)
		}
		total += n
	}
	return total, nil
}

func EncodeBool(e *Encoder, value bool) (int, error) {
	if value {
		e.scratch[0] = 1
	} else {
		e.scratch[0] = 0
	}
	return e.w.Write(e.scratch[:1])
}

func EncodeByte(e *Encoder, value byte) (int, error) {
	e.scratch[0] = value
	return e.w.Write(e.scratch[:1])
}

func EncodeUint16(e *Encoder, value uint16) (int, error) {
	binary.LittleEndian.PutUint16(e.scratch[:2], value)
	return e.w.Write(e.scratch[:2])
}

func EncodeUint32(e *Encoder, value uint32) (int, error) {
	binary.LittleEndian.PutUint32(e.scratch[:4], value)
	return e.w.Write(e.scratch[:4])
}

func EncodeUint64(e *Encoder, value uint64) (int, error) {
	binary.LittleEndian.PutUint64(e.scratch[:8], value)
	return e.w.Write(e.scratch[:8])
}

func encodeUint8(e *Encoder, v uint8) (int, error) {
	e.scratch[0] = byte(v)
	return e.w.Write(e.scratch[:1])
}

func encodeUint16(e *Encoder, v uint16) (int, error) {
	e.scratch[0] = byte(v)
	e.scratch[1] = byte(v >> 8)
	return e.w.Write(e.scratch[:2])
}

func encodeUint32(e *Encoder, v uint32) (int, error) {
	e.scratch[0] = byte(v)
	e.scratch[1] = byte(v >> 8)
	e.scratch[2] = byte(v >> 16)
	e.scratch[3] = byte(v >> 24)
	return e.w.Write(e.scratch[:4])
}

func encodeBigUint(e *Encoder, v uint64) (int, error) {
	needed := 8 - bits.LeadingZeros64(v)/8
	e.scratch[0] = byte(needed-4)<<2 | 0b11
	for i := 1; i <= needed; i++ {
		e.scratch[i] = byte(v)
		v >>= 8
	}
	return e.w.Write(e.scratch[:needed+1])
}

func EncodeCompact8(e *Encoder, v uint8) (int, error) {
	if v <= maxUint6 {
		return encodeUint8(e, v<<2)
	}
	return encodeUint16(e, uint16(v)<<2|0b01)
}

func EncodeCompact16(e *Encoder, v uint16) (int, error) {
	if v <= maxUint6 {
		return encodeUint8(e, uint8(v<<2))
	} else if v <= maxUint14 {
		return encodeUint16(e, v<<2|0b01)
	}
	return encodeUint32(e, uint32(v)<<2|0b10)
}

func EncodeCompact32(e *Encoder, v uint32) (int, error) {
	if v <= maxUint6 {
		return encodeUint8(e, uint8(v<<2))
	} else if v <= maxUint14 {
		return encodeUint16(e, uint16(v<<2|0b01))
	} else if v <= maxUint30 {
		return encodeUint32(e, v<<2|0b10)
	}
	return encodeBigUint(e, uint64(v))
}

func EncodeCompact64(e *Encoder, v uint64) (int, error) {
	if v <= maxUint6 {
		return encodeUint8(e, uint8(v<<2))
	} else if v <= maxUint14 {
		return encodeUint16(e, uint16(v<<2|0b01))
	} else if v <= maxUint30 {
		return encodeUint32(e, uint32(v<<2|0b10))
	}
	return encodeBigUint(e, uint64(v))
}

func EncodeLen(e *Encoder, v uint32, limit uint32) (int, error) {
	if v > limit {
		return 0, fmt.Errorf("%w: %d", ErrEncodeTooManyElements, limit)
	}
	return EncodeCompact32(e, v)
}

func EncodeOption(e *Encoder, value Encodable) (int, error) {
	if err := e.enterNested(); err != nil {
		return 0, err
	}
	defer e.leaveNested()
	if IsNil(value) {
		return EncodeBool(e, false)
	}
	total, err := EncodeBool(e, true)
	if err != nil {
		return 0, err
	}
	n, err := value.EncodeScale(e)
	if err != nil {
		return 0, err
	}
	return total + n, nil
}
func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}
