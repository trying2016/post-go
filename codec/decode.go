package codec

import (
	"errors"
	"fmt"
	"io"
)

const (
	maxUint6  = 1<<6 - 1
	maxUint8  = 1<<8 - 1
	maxUint14 = 1<<14 - 1
	maxUint16 = 1<<16 - 1
	maxUint30 = 1<<30 - 1
)

var (
	// ErrDecodeTooManyElements is returned when scale limit tag is used and collection has too many elements to decode.
	ErrDecodeTooManyElements = errors.New("too many elements to decode in collection with scale limit set")

	// ErrDecodeNestedTooDeep is returned when nested level is too deep.
	ErrDecodeNestedTooDeep = errors.New("nested level is too deep")
)

type Decodable interface {
	DecodeScale(*Decoder) (int, error)
}

type decoderOpts func(*Decoder)

// WithDecodeMaxNested sets the nested level of the decoder.
// A value of 0 means no nesting is allowed. The default value is 4.
func WithDecodeMaxNested(nested uint) decoderOpts {
	return func(d *Decoder) {
		d.maxNested = nested
	}
}

// WithDecodeMaxElements sets the maximum number of elements allowed in a collection.
// The default value is 1 << 20.
func WithDecodeMaxElements(elements uint32) decoderOpts {
	return func(e *Decoder) {
		e.maxElements = elements
	}
}

type Decoder struct {
	r           io.Reader
	scratch     [9]byte
	maxNested   uint
	maxElements uint32
}

func (d *Decoder) enterNested() error {
	if d.maxNested == 0 {
		return ErrDecodeNestedTooDeep
	}
	d.maxNested--
	return nil
}

func (e *Decoder) leaveNested() {
	e.maxNested++
}

func (d *Decoder) read(buf []byte) (int, error) {
	return io.ReadFull(d.r, buf)
}

func DecodeByte(d *Decoder) (byte, int, error) {
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return 0, n, err
	}
	return d.scratch[0], n, err
}

func DecodeCompact8(d *Decoder) (uint8, int, error) {
	var (
		value uint8
		total int
	)
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return value, 0, err
	}
	total += n
	switch d.scratch[0] % 4 {
	case 0:
		value = d.scratch[0] >> 2
	case 1:
		n, err := d.read(d.scratch[1:2])
		if err != nil {
			return value, 0, err
		}
		total += n
		rst := (uint16(d.scratch[0]) | uint16(d.scratch[1])<<8) >> 2
		if rst <= maxUint6 || rst > maxUint8 {
			return 0, 0, fmt.Errorf("value %d is out of range", rst)
		}
		value = uint8(rst)
	default:
		return 0, 0, fmt.Errorf("value will overflow uint8")
	}
	return value, total, nil
}

func DecodeCompact16(d *Decoder) (uint16, int, error) {
	var (
		value uint16
		total int
	)
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return value, 0, err
	}
	total += n
	switch d.scratch[0] % 4 {
	case 0:
		value = uint16(d.scratch[0]) >> 2
	case 1:
		n, err := d.read(d.scratch[1:2])
		if err != nil {
			return value, 0, err
		}
		total += n
		value = (uint16(d.scratch[0]) | uint16(d.scratch[1])<<8) >> 2
		if value <= maxUint6 || value > maxUint14 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	case 2:
		n, err := d.read(d.scratch[1:4])
		if err != nil {
			return value, 0, err
		}
		total += n
		rst := (uint32(d.scratch[0]) |
			uint32(d.scratch[1])<<8 |
			uint32(d.scratch[2])<<16 |
			uint32(d.scratch[3])<<24) >> 2
		if rst <= maxUint14 || rst > maxUint16 {
			return 0, 0, fmt.Errorf("value %d is out of range", rst)
		}
		value = uint16(rst)
	default:
		return 0, 0, fmt.Errorf("value will overflow uint16")
	}
	return value, total, nil
}

func DecodeCompact32(d *Decoder) (uint32, int, error) {
	var (
		value uint32
		total int
	)
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return value, 0, err
	}
	total += n
	switch d.scratch[0] % 4 {
	case 0:
		value = uint32(d.scratch[0]) >> 2
	case 1:
		n, err := d.read(d.scratch[1:2])
		if err != nil {
			return value, 0, err
		}
		total += n
		value = (uint32(d.scratch[0]) | uint32(d.scratch[1])<<8) >> 2
		if value <= maxUint6 || value > maxUint14 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	case 2:
		n, err := d.read(d.scratch[1:4])
		if err != nil {
			return value, 0, err
		}
		total += n
		value = (uint32(d.scratch[0]) |
			uint32(d.scratch[1])<<8 |
			uint32(d.scratch[2])<<16 |
			uint32(d.scratch[3])<<24) >> 2
		if value <= maxUint14 || value > maxUint30 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	case 3:
		needed := d.scratch[0]>>2 + 4
		if needed > 4 {
			return value, 0, fmt.Errorf("invalid compact32 needs %d bytes", needed)
		}
		_, err := d.read(d.scratch[:needed])
		if err != nil {
			return value, 0, err
		}
		total += int(needed)
		for i := 0; i < int(needed); i++ {
			value |= uint32(d.scratch[i]) << (8 * i)
		}
		if value <= maxUint30 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	}
	return value, total, nil
}

func DecodeLen(d *Decoder, limit uint32) (uint32, int, error) {
	v, n, err := DecodeCompact32(d)
	if err != nil {
		return v, n, err
	}
	if v > limit {
		return v, n, fmt.Errorf("%w: (%d > %d)", ErrDecodeTooManyElements, v, limit)
	}
	return v, n, err
}

func DecodeCompact64(d *Decoder) (uint64, int, error) {
	var (
		value uint64
		total int
	)
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return value, 0, fmt.Errorf("read compact header: %w", err)
	}
	total += n
	switch d.scratch[0] % 4 {
	case 0:
		value = uint64(d.scratch[0]) >> 2
	case 1:
		n, err := d.read(d.scratch[1:2])
		if err != nil {
			return 0, 0, err
		}
		total += n
		value = (uint64(d.scratch[0]) | uint64(d.scratch[1])<<8) >> 2
		if value <= maxUint6 || value > maxUint14 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	case 2:
		n, err := d.read(d.scratch[1:4])
		if err != nil {
			return 0, 0, err
		}
		total += n
		value = (uint64(d.scratch[0]) |
			uint64(d.scratch[1])<<8 |
			uint64(d.scratch[2])<<16 |
			uint64(d.scratch[3])<<24) >> 2
		if value <= maxUint14 || value > maxUint30 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	case 3:
		needed := d.scratch[0]>>2 + 4
		if needed > 8 {
			return value, 0, fmt.Errorf("invalid compact64 needs %d bytes", needed)
		}
		n, err := d.read(d.scratch[:needed])
		if err != nil {
			return 0, 0, err
		}
		total += n
		for i := 0; i < int(needed); i++ {
			value |= uint64(d.scratch[i]) << (8 * i)
		}
		if value <= maxUint30 {
			return 0, 0, fmt.Errorf("value %d is out of range", value)
		}
	}
	return value, total, nil
}

func DecodeBool(d *Decoder) (bool, int, error) {
	n, err := d.read(d.scratch[:1])
	if err != nil {
		return false, 0, err
	}
	if d.scratch[0] == 1 {
		return true, n, nil
	}
	return false, n, nil
}

func DecodeByteSlice(d *Decoder) ([]byte, int, error) {
	return DecodeByteSliceWithLimit(d, d.maxElements)
}

func DecodeByteSliceWithLimit(d *Decoder, limit uint32) ([]byte, int, error) {
	lth, total, err := DecodeLen(d, limit)
	if err != nil {
		return nil, 0, err
	}
	if lth == 0 {
		return nil, total, nil
	}
	value := make([]byte, lth)
	n, err := DecodeByteArray(d, value)
	if err != nil {
		return value, 0, err
	}
	return value, total + n, nil
}

func DecodeByteArray(d *Decoder, value []byte) (int, error) {
	return d.read(value)
}

func DecodeString(d *Decoder) (string, int, error) {
	return DecodeStringWithLimit(d, d.maxElements)
}

func DecodeStringWithLimit(d *Decoder, limit uint32) (string, int, error) {
	bytes, total, err := DecodeByteSliceWithLimit(d, limit)
	return string(bytes), total, err
}

func DecodeSliceOfByteSlice(d *Decoder) ([][]byte, int, error) {
	return DecodeSliceOfByteSliceWithLimit(d, d.maxElements)
}

func DecodeSliceOfByteSliceWithLimit(d *Decoder, limit uint32) ([][]byte, int, error) {
	resultLen, total, err := DecodeLen(d, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("DecodeLen failed: %w", err)
	}
	if resultLen == 0 {
		return nil, 0, nil
	}
	result := make([][]byte, 0, resultLen)

	for i := uint32(0); i < resultLen; i++ {
		val, n, err := DecodeByteSlice(d)
		if err != nil {
			return nil, 0, fmt.Errorf("DecodeByteSlice failed: %w", err)
		}
		result = append(result, val)
		total += n
	}

	return result, total, nil
}

func DecodeStringSlice(d *Decoder) ([]string, int, error) {
	return DecodeStringSliceWithLimit(d, d.maxElements)
}

func DecodeStringSliceWithLimit(d *Decoder, limit uint32) ([]string, int, error) {
	sliceOfByteSlices, n, err := DecodeSliceOfByteSliceWithLimit(d, limit)
	if err != nil {
		return nil, 0, err
	}
	if sliceOfByteSlices == nil {
		return nil, 0, nil
	}
	result := make([]string, 0, len(sliceOfByteSlices))
	for i := range sliceOfByteSlices {
		result = append(result, bytesToString(sliceOfByteSlices[i]))
	}

	return result, n, nil
}

func DecodeOption(d *Decoder, h Decodable) (bool, int, error) {
	if err := d.enterNested(); err != nil {
		return false, 0, err
	}
	defer d.leaveNested()
	exists, total, err := DecodeBool(d)
	if !exists || err != nil {
		return false, total, err
	}
	n, err := h.DecodeScale(d)
	if err != nil {
		return false, 0, err
	}
	return true, total + n, nil
}
