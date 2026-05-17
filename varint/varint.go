// Package varint provides variable-length integer encoding and decoding.
//
// Unsigned integers use the standard unsigned varint encoding from
// encoding/binary. Signed integers use ZigZag encoding so that small
// negative values (e.g. -1) encode into one or two bytes rather than
// ten.
package varint

import (
	"encoding/binary"
	"errors"
)

// ErrOverflow is returned when a decoded value overflows the target type.
var ErrOverflow = errors.New("varint: overflow")

// ErrIncomplete is returned when src does not contain a complete varint.
var ErrIncomplete = errors.New("varint: incomplete")

// EncodeU64 appends the unsigned varint encoding of v to dst and returns
// the number of bytes written. dst must have at least SizeU64(v) bytes
// of capacity remaining; callers typically pass a pre-allocated slice.
func EncodeU64(dst []byte, v uint64) int {
	return binary.PutUvarint(dst, v)
}

// DecodeU64 decodes an unsigned varint from src. It returns the value,
// the number of bytes consumed, and any error.
func DecodeU64(src []byte) (value uint64, n int, err error) {
	v, n := binary.Uvarint(src)
	switch {
	case n == 0:
		return 0, 0, ErrIncomplete
	case n < 0:
		return 0, 0, ErrOverflow
	}
	return v, n, nil
}

// EncodeI64 appends the ZigZag-encoded signed varint of v to dst and
// returns the number of bytes written.
func EncodeI64(dst []byte, v int64) int {
	return binary.PutVarint(dst, v)
}

// DecodeI64 decodes a ZigZag-encoded signed varint from src. It returns
// the value, the number of bytes consumed, and any error.
func DecodeI64(src []byte) (value int64, n int, err error) {
	v, n := binary.Varint(src)
	switch {
	case n == 0:
		return 0, 0, ErrIncomplete
	case n < 0:
		return 0, 0, ErrOverflow
	}
	return v, n, nil
}

// SizeU64 returns the number of bytes required to encode v as an
// unsigned varint.
func SizeU64(v uint64) int {
	switch {
	case v < 1<<7:
		return 1
	case v < 1<<14:
		return 2
	case v < 1<<21:
		return 3
	case v < 1<<28:
		return 4
	case v < 1<<35:
		return 5
	case v < 1<<42:
		return 6
	case v < 1<<49:
		return 7
	case v < 1<<56:
		return 8
	case v < 1<<63:
		return 9
	default:
		return 10
	}
}

// SizeI64 returns the number of bytes required to encode v as a
// ZigZag-encoded signed varint.
func SizeI64(v int64) int {
	// ZigZag maps v to (v << 1) ^ (v >> 63).
	var u uint64
	if v >= 0 {
		u = uint64(v) << 1
	} else {
		u = ^(uint64(v) << 1)
	}
	return SizeU64(u)
}
