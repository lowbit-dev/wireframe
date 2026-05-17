package frame

import (
	"encoding/binary"

	"lowbit.dev/wireframe/varint"
)

// Field represents one named slot in the frame header. It reads from
// and writes to a shared Values, which is how values are passed between
// the caller and the codec.
type Field interface {
	// Name returns the field's string key used in Values.
	Name() string
	// Size returns the byte width of this field on the wire.
	// For PayloadLength the size depends on the length encoding; for
	// fixed-width fields it is always a constant.
	Size() int
	// Encode writes the field value from vals into dst and returns the
	// number of bytes written.
	Encode(dst []byte, vals *Values) (int, error)
	// Decode reads the field value from src into vals and returns the
	// number of bytes consumed.
	Decode(src []byte, vals *Values) (int, error)
}

// --- uint8 field ---

type uint8Field struct{ name string }

// Uint8 returns a field that encodes and decodes a single byte.
func Uint8(name string) Field { return uint8Field{name} }

func (f uint8Field) Name() string { return f.name }
func (f uint8Field) Size() int    { return 1 }

func (f uint8Field) Encode(dst []byte, vals *Values) (int, error) {
	dst[0] = vals.Uint8(f.name)
	return 1, nil
}

func (f uint8Field) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, src[0])
	return 1, nil
}

// --- uint16 big-endian field ---

type uint16BEField struct{ name string }

// Uint16BE returns a 2-byte big-endian unsigned integer field.
func Uint16BE(name string) Field { return uint16BEField{name} }

func (f uint16BEField) Name() string { return f.name }
func (f uint16BEField) Size() int    { return 2 }

func (f uint16BEField) Encode(dst []byte, vals *Values) (int, error) {
	binary.BigEndian.PutUint16(dst, vals.Uint16(f.name))
	return 2, nil
}

func (f uint16BEField) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, binary.BigEndian.Uint16(src))
	return 2, nil
}

// --- uint32 big-endian field ---

type uint32BEField struct{ name string }

// Uint32BE returns a 4-byte big-endian unsigned integer field.
func Uint32BE(name string) Field { return uint32BEField{name} }

func (f uint32BEField) Name() string { return f.name }
func (f uint32BEField) Size() int    { return 4 }

func (f uint32BEField) Encode(dst []byte, vals *Values) (int, error) {
	binary.BigEndian.PutUint32(dst, vals.Uint32(f.name))
	return 4, nil
}

func (f uint32BEField) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, binary.BigEndian.Uint32(src))
	return 4, nil
}

// --- PayloadLength field ---

// payloadLengthEncoding determines how the length value is stored on
// the wire. The codec chooses based on Format.MaxPayload at construction.
type payloadLengthEncoding uint8

const (
	plEncUint8  payloadLengthEncoding = iota // 0–255
	plEncUint16                              // 0–65535
	plEncUint32                              // 0–4294967295
	plEncVarint                              // variable length
)

// payloadLengthField writes the payload byte count during encoding and
// reads the expected byte count during decoding.
type payloadLengthField struct {
	name string
	enc  payloadLengthEncoding
}

// PayloadLength returns a field that encodes/decodes the payload byte
// count. The encoding width is selected automatically from Format.MaxPayload
// at validation time; before then it defaults to uint32.
func PayloadLength(name string) Field {
	return payloadLengthField{name: name, enc: plEncUint32}
}

func (f payloadLengthField) Name() string { return f.name }

func (f payloadLengthField) Size() int {
	switch f.enc {
	case plEncUint8:
		return 1
	case plEncUint16:
		return 2
	case plEncVarint:
		return -1 // variable
	default:
		return 4
	}
}

func (f payloadLengthField) Encode(dst []byte, vals *Values) (int, error) {
	length := vals.Uint32(f.name)
	switch f.enc {
	case plEncUint8:
		dst[0] = uint8(length)
		return 1, nil
	case plEncUint16:
		binary.BigEndian.PutUint16(dst, uint16(length))
		return 2, nil
	case plEncVarint:
		n := varint.EncodeU64(dst, uint64(length))
		return n, nil
	default:
		binary.BigEndian.PutUint32(dst, length)
		return 4, nil
	}
}

func (f payloadLengthField) Decode(src []byte, vals *Values) (int, error) {
	switch f.enc {
	case plEncUint8:
		Set(vals, f.name, uint32(src[0]))
		return 1, nil
	case plEncUint16:
		Set(vals, f.name, uint32(binary.BigEndian.Uint16(src)))
		return 2, nil
	case plEncVarint:
		v, n, err := varint.DecodeU64(src)
		if err != nil {
			return 0, err
		}
		Set(vals, f.name, uint32(v))
		return n, nil
	default:
		Set(vals, f.name, binary.BigEndian.Uint32(src))
		return 4, nil
	}
}
