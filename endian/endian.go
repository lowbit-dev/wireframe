// Package endian provides type-safe helpers for reading and writing
// fixed-width unsigned integers with explicit byte order.
//
// Using generics it collapses what would otherwise be twelve separate
// functions (four types × three byte orders) into four.
//
// User-defined types built on unsigned integer bases satisfy the
// Unsigned constraint and work without explicit casting:
//
//	type SessionID uint32
//	endian.PutBE(buf, SessionID(42))
package endian

import "encoding/binary"

// Unsigned is the constraint satisfied by all unsigned integer types
// as well as user-defined types with an unsigned integer base.
type Unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// PutBE encodes v into dst in big-endian byte order.
// dst must be at least as long as the byte size of T.
func PutBE[T Unsigned](dst []byte, v T) {
	switch any(v).(type) {
	case uint8:
		dst[0] = uint8(v)
	case uint16:
		binary.BigEndian.PutUint16(dst, uint16(v))
	case uint32:
		binary.BigEndian.PutUint32(dst, uint32(v))
	case uint64:
		binary.BigEndian.PutUint64(dst, uint64(v))
	default:
		putBEBySize(dst, v)
	}
}

// ReadBE decodes a value of type T from src in big-endian byte order.
// src must be at least as long as the byte size of T.
func ReadBE[T Unsigned](src []byte) T {
	switch any(T(0)).(type) {
	case uint8:
		return T(src[0])
	case uint16:
		return T(binary.BigEndian.Uint16(src))
	case uint32:
		return T(binary.BigEndian.Uint32(src))
	case uint64:
		return T(binary.BigEndian.Uint64(src))
	default:
		return readBEBySize[T](src)
	}
}

// PutLE encodes v into dst in little-endian byte order.
// dst must be at least as long as the byte size of T.
func PutLE[T Unsigned](dst []byte, v T) {
	switch any(v).(type) {
	case uint8:
		dst[0] = uint8(v)
	case uint16:
		binary.LittleEndian.PutUint16(dst, uint16(v))
	case uint32:
		binary.LittleEndian.PutUint32(dst, uint32(v))
	case uint64:
		binary.LittleEndian.PutUint64(dst, uint64(v))
	default:
		putLEBySize(dst, v)
	}
}

// ReadLE decodes a value of type T from src in little-endian byte order.
// src must be at least as long as the byte size of T.
func ReadLE[T Unsigned](src []byte) T {
	switch any(T(0)).(type) {
	case uint8:
		return T(src[0])
	case uint16:
		return T(binary.LittleEndian.Uint16(src))
	case uint32:
		return T(binary.LittleEndian.Uint32(src))
	case uint64:
		return T(binary.LittleEndian.Uint64(src))
	default:
		return readLEBySize[T](src)
	}
}

// putBEBySize handles user-defined types by inspecting the size at runtime.
func putBEBySize[T Unsigned](dst []byte, v T) {
	u := uint64(v)
	switch sizeOf[T]() {
	case 1:
		dst[0] = byte(u)
	case 2:
		binary.BigEndian.PutUint16(dst, uint16(u))
	case 4:
		binary.BigEndian.PutUint32(dst, uint32(u))
	case 8:
		binary.BigEndian.PutUint64(dst, u)
	}
}

func readBEBySize[T Unsigned](src []byte) T {
	switch sizeOf[T]() {
	case 1:
		return T(src[0])
	case 2:
		return T(binary.BigEndian.Uint16(src))
	case 4:
		return T(binary.BigEndian.Uint32(src))
	case 8:
		return T(binary.BigEndian.Uint64(src))
	}
	return 0
}

func putLEBySize[T Unsigned](dst []byte, v T) {
	u := uint64(v)
	switch sizeOf[T]() {
	case 1:
		dst[0] = byte(u)
	case 2:
		binary.LittleEndian.PutUint16(dst, uint16(u))
	case 4:
		binary.LittleEndian.PutUint32(dst, uint32(u))
	case 8:
		binary.LittleEndian.PutUint64(dst, u)
	}
}

func readLEBySize[T Unsigned](src []byte) T {
	switch sizeOf[T]() {
	case 1:
		return T(src[0])
	case 2:
		return T(binary.LittleEndian.Uint16(src))
	case 4:
		return T(binary.LittleEndian.Uint32(src))
	case 8:
		return T(binary.LittleEndian.Uint64(src))
	}
	return 0
}

// sizeOf returns the byte width of T by inspecting a zero value.
func sizeOf[T Unsigned]() int {
	var zero T
	switch any(zero).(type) {
	case uint8:
		return 1
	case uint16:
		return 2
	case uint32:
		return 4
	case uint64:
		return 8
	}
	// Fallback for user-defined types: use unsafe-free size detection via
	// bit shifting. We know T is at most uint64 wide.
	var v T = ^T(0)
	switch {
	case v>>8 == 0:
		return 1
	case v>>16 == 0:
		return 2
	case v>>32 == 0:
		return 4
	default:
		return 8
	}
}
