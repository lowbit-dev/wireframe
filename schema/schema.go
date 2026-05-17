// Package schema describes binary payload layouts.
//
// Where frame handles packet boundaries and header metadata, schema
// handles the structured data inside the payload. The two packages are
// independent and compose naturally.
//
// No reflection, no struct tags, no automatic field discovery. Values
// are mapped explicitly by the caller.
package schema

import (
	"encoding/binary"
	"errors"
	"fmt"

	"lowbit.dev/wireframe/varint"
)

// Sentinel errors.
var (
	ErrIncomplete  = errors.New("schema: incomplete")
	ErrInvalidData = errors.New("schema: invalid data")
	ErrOverflow    = errors.New("schema: overflow")
)

// --- Values (mirrors frame.Values, but scoped to schema) ---

const maxInlineFields = 16

type rawValue struct {
	u64 uint64
	b   []byte
}

// Values carries payload field values between the caller and the codec.
type Values struct {
	keys     [maxInlineFields]string
	vals     [maxInlineFields]rawValue
	n        int
	overflow map[string]rawValue
}

// NewValues allocates a new Values on the heap.
func NewValues() *Values { return &Values{} }

// Reset clears all stored values.
func (v *Values) Reset() {
	v.n = 0
	v.overflow = nil
}

func (v *Values) set(key string, rv rawValue) {
	for i := 0; i < v.n; i++ {
		if v.keys[i] == key {
			v.vals[i] = rv
			return
		}
	}
	if v.overflow != nil {
		if _, ok := v.overflow[key]; ok {
			v.overflow[key] = rv
			return
		}
	}
	if v.n < maxInlineFields {
		v.keys[v.n] = key
		v.vals[v.n] = rv
		v.n++
		return
	}
	if v.overflow == nil {
		v.overflow = make(map[string]rawValue)
	}
	v.overflow[key] = rv
}

func (v *Values) get(key string) (rawValue, bool) {
	for i := 0; i < v.n; i++ {
		if v.keys[i] == key {
			return v.vals[i], true
		}
	}
	if v.overflow != nil {
		rv, ok := v.overflow[key]
		return rv, ok
	}
	return rawValue{}, false
}

// Uint8 returns the uint8 value stored under key, or 0 if absent.
func (v *Values) Uint8(key string) uint8 {
	rv, _ := v.get(key)
	return uint8(rv.u64)
}

// Uint16 returns the uint16 value stored under key, or 0 if absent.
func (v *Values) Uint16(key string) uint16 {
	rv, _ := v.get(key)
	return uint16(rv.u64)
}

// Uint32 returns the uint32 value stored under key, or 0 if absent.
func (v *Values) Uint32(key string) uint32 {
	rv, _ := v.get(key)
	return uint32(rv.u64)
}

// Bytes returns the []byte value stored under key, or nil if absent.
func (v *Values) Bytes(key string) []byte {
	rv, _ := v.get(key)
	return rv.b
}

// Settable is the constraint for values that can be stored in Values.
type Settable interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~bool | ~[]byte
}

// Set stores v under key in vals.
func Set[T Settable](vals *Values, key string, v T) {
	var rv rawValue
	switch x := any(v).(type) {
	case uint8:
		rv.u64 = uint64(x)
	case uint16:
		rv.u64 = uint64(x)
	case uint32:
		rv.u64 = uint64(x)
	case uint64:
		rv.u64 = x
	case bool:
		if x {
			rv.u64 = 1
		}
	case []byte:
		rv.b = x
		vals.set(key, rv)
		return
	default:
		setUnsigned(vals, key, v)
		return
	}
	vals.set(key, rv)
}

func setUnsigned[T Settable](vals *Values, key string, v T) {
	var u uint64
	switch x := any(v).(type) {
	case uint8:
		u = uint64(x)
	case uint16:
		u = uint64(x)
	case uint32:
		u = uint64(x)
	case uint64:
		u = x
	case bool:
		if x {
			u = 1
		}
	case []byte:
		vals.set(key, rawValue{b: x})
		return
	}
	vals.set(key, rawValue{u64: u})
}

// --- Field interface ---

// Field represents one named slot in a payload schema.
type Field interface {
	Name() string
	// Size returns the byte size of this field given the current vals,
	// or -1 for variable-length fields.
	Size(vals *Values) int
	Encode(dst []byte, vals *Values) (int, error)
	Decode(src []byte, vals *Values) (int, error)
}

// --- uint8 ---

type uint8Field struct{ name string }

// Uint8 returns a 1-byte unsigned integer payload field.
func Uint8(name string) Field { return uint8Field{name} }

func (f uint8Field) Name() string       { return f.name }
func (f uint8Field) Size(_ *Values) int { return 1 }
func (f uint8Field) Encode(dst []byte, vals *Values) (int, error) {
	dst[0] = vals.Uint8(f.name)
	return 1, nil
}
func (f uint8Field) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, src[0])
	return 1, nil
}

// --- uint16 BE ---

type uint16BEField struct{ name string }

// Uint16BE returns a 2-byte big-endian unsigned integer payload field.
func Uint16BE(name string) Field { return uint16BEField{name} }

func (f uint16BEField) Name() string       { return f.name }
func (f uint16BEField) Size(_ *Values) int { return 2 }
func (f uint16BEField) Encode(dst []byte, vals *Values) (int, error) {
	binary.BigEndian.PutUint16(dst, vals.Uint16(f.name))
	return 2, nil
}
func (f uint16BEField) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, binary.BigEndian.Uint16(src))
	return 2, nil
}

// --- uint32 BE ---

type uint32BEField struct{ name string }

// Uint32BE returns a 4-byte big-endian unsigned integer payload field.
func Uint32BE(name string) Field { return uint32BEField{name} }

func (f uint32BEField) Name() string       { return f.name }
func (f uint32BEField) Size(_ *Values) int { return 4 }
func (f uint32BEField) Encode(dst []byte, vals *Values) (int, error) {
	binary.BigEndian.PutUint32(dst, vals.Uint32(f.name))
	return 4, nil
}
func (f uint32BEField) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, binary.BigEndian.Uint32(src))
	return 4, nil
}

// --- bool ---

type boolField struct{ name string }

// Bool returns a 1-byte boolean payload field (0 = false, non-zero = true).
func Bool(name string) Field { return boolField{name} }

func (f boolField) Name() string       { return f.name }
func (f boolField) Size(_ *Values) int { return 1 }
func (f boolField) Encode(dst []byte, vals *Values) (int, error) {
	rv, _ := vals.get(f.name)
	if rv.u64 != 0 {
		dst[0] = 1
	} else {
		dst[0] = 0
	}
	return 1, nil
}
func (f boolField) Decode(src []byte, vals *Values) (int, error) {
	Set(vals, f.name, src[0] != 0)
	return 1, nil
}

// --- varint ---

type varintField struct{ name string }

// Varint returns a variable-length unsigned integer payload field.
func Varint(name string) Field { return varintField{name} }

func (f varintField) Name() string { return f.name }
func (f varintField) Size(vals *Values) int {
	v := vals.Uint32(f.name)
	return varint.SizeU64(uint64(v))
}
func (f varintField) Encode(dst []byte, vals *Values) (int, error) {
	return varint.EncodeU64(dst, uint64(vals.Uint32(f.name))), nil
}
func (f varintField) Decode(src []byte, vals *Values) (int, error) {
	v, n, err := varint.DecodeU64(src)
	if err != nil {
		return 0, err
	}
	Set(vals, f.name, uint32(v))
	return n, nil
}

// --- fixed-length bytes ---

type bytesField struct {
	name   string
	length int
}

// Bytes returns a fixed-length byte slice payload field.
func Bytes(name string, length int) Field { return bytesField{name, length} }

func (f bytesField) Name() string       { return f.name }
func (f bytesField) Size(_ *Values) int { return f.length }
func (f bytesField) Encode(dst []byte, vals *Values) (int, error) {
	b := vals.Bytes(f.name)
	if len(b) != f.length {
		return 0, fmt.Errorf("%w: field %q: expected %d bytes, got %d", ErrInvalidData, f.name, f.length, len(b))
	}
	copy(dst, b)
	return f.length, nil
}
func (f bytesField) Decode(src []byte, vals *Values) (int, error) {
	if len(src) < f.length {
		return 0, ErrIncomplete
	}
	cp := make([]byte, f.length)
	copy(cp, src[:f.length])
	Set(vals, f.name, cp)
	return f.length, nil
}

// --- variable-length bytes (length-prefixed) ---

type varBytesField struct {
	name        string
	lengthField string
}

// VarBytes returns a variable-length byte slice field whose length is
// stored in a preceding field named lengthField.
func VarBytes(name, lengthField string) Field { return varBytesField{name, lengthField} }

func (f varBytesField) Name() string { return f.name }
func (f varBytesField) Size(vals *Values) int {
	return int(vals.Uint32(f.lengthField))
}
func (f varBytesField) Encode(dst []byte, vals *Values) (int, error) {
	b := vals.Bytes(f.name)
	// Write the length into the length field value before encoding the body.
	Set(vals, f.lengthField, uint32(len(b)))
	copy(dst, b)
	return len(b), nil
}
func (f varBytesField) Decode(src []byte, vals *Values) (int, error) {
	length := int(vals.Uint32(f.lengthField))
	if len(src) < length {
		return 0, ErrIncomplete
	}
	cp := make([]byte, length)
	copy(cp, src[:length])
	Set(vals, f.name, cp)
	return length, nil
}

// varBytesAutoField is a self-contained length-prefixed byte slice field.
// It encodes as [uint16 big-endian length][data...] and does not require a
// separate length field in the schema. Used by the typed API.
type varBytesAutoField struct{ name string }

// VarBytesAuto returns a variable-length byte slice field with a 2-byte
// big-endian length prefix. The length field is managed internally and
// does not appear in Values.
func VarBytesAuto(name string) Field { return varBytesAutoField{name} }

func (f varBytesAutoField) Name() string { return f.name }
func (f varBytesAutoField) Size(vals *Values) int {
	b := vals.Bytes(f.name)
	return 2 + len(b)
}
func (f varBytesAutoField) Encode(dst []byte, vals *Values) (int, error) {
	b := vals.Bytes(f.name)
	binary.BigEndian.PutUint16(dst, uint16(len(b)))
	copy(dst[2:], b)
	return 2 + len(b), nil
}
func (f varBytesAutoField) Decode(src []byte, vals *Values) (int, error) {
	if len(src) < 2 {
		return 0, ErrIncomplete
	}
	length := int(binary.BigEndian.Uint16(src))
	if len(src) < 2+length {
		return 0, ErrIncomplete
	}
	cp := make([]byte, length)
	copy(cp, src[2:2+length])
	Set(vals, f.name, cp)
	return 2 + length, nil
}

// --- Schema ---

// Schema describes the binary layout of a payload. It encodes and
// decodes named fields in the order they were declared.
type Schema struct {
	Fields []Field
}

// New constructs a Schema from the given fields.
func New(fields ...Field) *Schema {
	return &Schema{Fields: fields}
}

// Encode serializes the values in vals according to the schema, appending
// the result to dst.
func (s *Schema) Encode(dst []byte, vals *Values) ([]byte, error) {
	// Pre-pass: for VarBytes fields (with external length fields), populate
	// the length field value so encoding proceeds in order.
	for _, f := range s.Fields {
		if vb, ok := f.(varBytesField); ok {
			b := vals.Bytes(vb.name)
			Set(vals, vb.lengthField, uint32(len(b)))
		}
	}

	var scratch [512]byte
	for _, f := range s.Fields {
		sz := f.Size(vals)
		if sz >= 0 {
			if sz > len(scratch) {
				// Field too large for scratch — allocate.
				buf := make([]byte, sz)
				n, err := f.Encode(buf, vals)
				if err != nil {
					return nil, err
				}
				dst = append(dst, buf[:n]...)
			} else {
				dst = append(dst, scratch[:sz]...)
				_, err := f.Encode(dst[len(dst)-sz:], vals)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// Variable-length field (varint).
			var tmp [10]byte
			n, err := f.Encode(tmp[:], vals)
			if err != nil {
				return nil, err
			}
			dst = append(dst, tmp[:n]...)
		}
	}
	return dst, nil
}

// Decode deserializes fields from src into vals. Returns the number of
// bytes consumed.
func (s *Schema) Decode(src []byte, vals *Values) (int, error) {
	pos := 0
	for _, f := range s.Fields {
		minBytes := f.Size(vals)
		if minBytes < 0 {
			minBytes = 1
		}
		if len(src)-pos < minBytes {
			return 0, ErrIncomplete
		}
		n, err := f.Decode(src[pos:], vals)
		if err != nil {
			return 0, err
		}
		pos += n
	}
	return pos, nil
}
