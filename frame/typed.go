package frame

import "lowbit.dev/wireframe/checksum"

// Unsigned is the constraint satisfied by all unsigned integer types and
// user-defined types with an unsigned integer base. It is used by
// PayloadLengthField to allow binding the length field to any integer
// width in the caller's header struct.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// FormatOption[S] is a functional option for NewTypedFormat.
type FormatOption[S any] func(f *typedFormatConfig[S])

// typedFormatConfig accumulates options before building the TypedFormat.
type typedFormatConfig[S any] struct {
	prefix     []byte
	fields     []typedField[S]
	checksum   checksum.Algorithm
	suffix     []byte
	maxPayload int
}

// typedField pairs an untyped Field with the accessor that bridges
// between the wire representation and a struct field in S.
type typedField[S any] struct {
	field  Field
	encode func(f Field, vals *Values, v *S) error
	decode func(f Field, vals *Values, v *S)
}

// WithPrefix sets the frame prefix bytes.
func WithPrefix[S any](b []byte) FormatOption[S] {
	return func(c *typedFormatConfig[S]) { c.prefix = b }
}

// WithChecksum sets the checksum algorithm.
func WithChecksum[S any](alg checksum.Algorithm) FormatOption[S] {
	return func(c *typedFormatConfig[S]) { c.checksum = alg }
}

// WithMaxPayload sets the maximum payload size.
func WithMaxPayload[S any](n int) FormatOption[S] {
	return func(c *typedFormatConfig[S]) { c.maxPayload = n }
}

// Uint8Field adds a uint8 header field bound to a struct field via acc.
func Uint8Field[S any](name string, acc func(*S) *uint8) FormatOption[S] {
	f := Uint8(name)
	return func(c *typedFormatConfig[S]) {
		c.fields = append(c.fields, typedField[S]{
			field: f,
			encode: func(field Field, vals *Values, v *S) error {
				Set(vals, field.Name(), *acc(v))
				return nil
			},
			decode: func(field Field, vals *Values, v *S) {
				*acc(v) = vals.Uint8(field.Name())
			},
		})
	}
}

// Uint16BEField adds a uint16 big-endian header field bound to a struct field.
func Uint16BEField[S any](name string, acc func(*S) *uint16) FormatOption[S] {
	f := Uint16BE(name)
	return func(c *typedFormatConfig[S]) {
		c.fields = append(c.fields, typedField[S]{
			field: f,
			encode: func(field Field, vals *Values, v *S) error {
				Set(vals, field.Name(), *acc(v))
				return nil
			},
			decode: func(field Field, vals *Values, v *S) {
				*acc(v) = vals.Uint16(field.Name())
			},
		})
	}
}

// Uint32BEField adds a uint32 big-endian header field bound to a struct field.
func Uint32BEField[S any](name string, acc func(*S) *uint32) FormatOption[S] {
	f := Uint32BE(name)
	return func(c *typedFormatConfig[S]) {
		c.fields = append(c.fields, typedField[S]{
			field: f,
			encode: func(field Field, vals *Values, v *S) error {
				Set(vals, field.Name(), *acc(v))
				return nil
			},
			decode: func(field Field, vals *Values, v *S) {
				*acc(v) = vals.Uint32(field.Name())
			},
		})
	}
}

// PayloadLengthField adds a payload-length field bound to a struct field.
// T must be an unsigned integer type; the accessor returns a pointer to it.
func PayloadLengthField[S any, T Unsigned](name string, acc func(*S) *T) FormatOption[S] {
	f := PayloadLength(name)
	return func(c *typedFormatConfig[S]) {
		c.fields = append(c.fields, typedField[S]{
			field: f,
			encode: func(field Field, vals *Values, v *S) error {
				// PayloadLength is auto-populated by Format.Encode; we write
				// it back to the struct after decoding only.
				return nil
			},
			decode: func(field Field, vals *Values, v *S) {
				*acc(v) = T(vals.Uint32(field.Name()))
			},
		})
	}
}

// TypedFormat[T] binds a frame layout to a concrete header struct type.
// It wraps the untyped Format and translates between *T and *Values using
// the accessor functions supplied to NewTypedFormat.
type TypedFormat[T any] struct {
	format *Format
	fields []typedField[T]
}

// NewTypedFormat constructs a TypedFormat from the given options.
func NewTypedFormat[T any](opts ...FormatOption[T]) *TypedFormat[T] {
	cfg := &typedFormatConfig[T]{maxPayload: 64 * 1024}
	for _, opt := range opts {
		opt(cfg)
	}

	rawFields := make([]Field, len(cfg.fields))
	for i, tf := range cfg.fields {
		rawFields[i] = tf.field
	}

	return &TypedFormat[T]{
		format: &Format{
			Prefix:     cfg.prefix,
			Header:     rawFields,
			Checksum:   cfg.checksum,
			Suffix:     cfg.suffix,
			MaxPayload: cfg.maxPayload,
		},
		fields: cfg.fields,
	}
}

// Format returns the underlying untyped Format.
func (tf *TypedFormat[T]) Format() *Format {
	return tf.format
}

// Encode encodes payload into a frame, populating header fields from v.
// PayloadLength is handled automatically and does not need to be set in v.
func (tf *TypedFormat[T]) Encode(dst []byte, payload []byte, v *T) ([]byte, error) {
	var vals Values
	for _, tf := range tf.fields {
		if err := tf.encode(tf.field, &vals, v); err != nil {
			return nil, err
		}
	}
	return tf.format.Encode(dst, payload, &vals)
}

// Decode decodes a single frame from src, populating v with the header
// field values. Returns the payload, the number of bytes consumed from
// src, and any error.
func (tf *TypedFormat[T]) Decode(src []byte, v *T) (payload []byte, n int, err error) {
	payload, vals, n, err := tf.format.Decode(src)
	if err != nil {
		return nil, 0, err
	}
	tf.FillHeader(vals, v)
	return payload, n, nil
}

// FillHeader populates v from the values in vals using the accessor
// functions bound at construction. This is the decode-from-values path
// used by stream.TypedReader.
func (tf *TypedFormat[T]) FillHeader(vals *Values, v *T) {
	for _, f := range tf.fields {
		f.decode(f.field, vals, v)
	}
}
