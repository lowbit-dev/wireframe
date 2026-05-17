package schema

// SchemaOption[S] is a functional option for NewTyped.
type SchemaOption[S any] func(c *typedSchemaConfig[S])

// typedSchemaConfig accumulates options before building the TypedSchema.
type typedSchemaConfig[S any] struct {
	fields []typedField[S]
}

// typedField pairs a schema Field with accessors for encoding and decoding.
type typedField[S any] struct {
	field  Field
	encode func(f Field, vals *Values, v *S) error
	decode func(f Field, vals *Values, v *S)
}

// Uint8Field adds a uint8 payload field bound to a struct field via acc.
func Uint8Field[S any](name string, acc func(*S) *uint8) SchemaOption[S] {
	f := Uint8(name)
	return func(c *typedSchemaConfig[S]) {
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

// Uint16BEField adds a uint16 big-endian payload field.
func Uint16BEField[S any](name string, acc func(*S) *uint16) SchemaOption[S] {
	f := Uint16BE(name)
	return func(c *typedSchemaConfig[S]) {
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

// Uint32BEField adds a uint32 big-endian payload field.
func Uint32BEField[S any](name string, acc func(*S) *uint32) SchemaOption[S] {
	f := Uint32BE(name)
	return func(c *typedSchemaConfig[S]) {
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

// VarBytesField adds a variable-length byte slice field. The length is
// automatically encoded as a 2-byte big-endian prefix and does not
// require a separate length field in the struct.
func VarBytesField[S any](name string, acc func(*S) *[]byte) SchemaOption[S] {
	f := VarBytesAuto(name)
	return func(c *typedSchemaConfig[S]) {
		c.fields = append(c.fields, typedField[S]{
			field: f,
			encode: func(field Field, vals *Values, v *S) error {
				Set(vals, field.Name(), *acc(v))
				return nil
			},
			decode: func(field Field, vals *Values, v *S) {
				*acc(v) = vals.Bytes(field.Name())
			},
		})
	}
}

// TypedSchema[T] binds a payload layout to a concrete struct type T.
type TypedSchema[T any] struct {
	schema *Schema
	fields []typedField[T]
}

// NewTyped constructs a TypedSchema from the given options.
func NewTyped[T any](opts ...SchemaOption[T]) *TypedSchema[T] {
	cfg := &typedSchemaConfig[T]{}
	for _, opt := range opts {
		opt(cfg)
	}
	rawFields := make([]Field, len(cfg.fields))
	for i, tf := range cfg.fields {
		rawFields[i] = tf.field
	}
	return &TypedSchema[T]{
		schema: New(rawFields...),
		fields: cfg.fields,
	}
}

// Encode serializes v into a binary payload, appending the result to dst.
func (ts *TypedSchema[T]) Encode(dst []byte, v *T) ([]byte, error) {
	var vals Values
	for _, f := range ts.fields {
		if err := f.encode(f.field, &vals, v); err != nil {
			return nil, err
		}
	}
	return ts.schema.Encode(dst, &vals)
}

// Decode deserializes a binary payload into v. Returns the number of
// bytes consumed.
func (ts *TypedSchema[T]) Decode(src []byte, v *T) (int, error) {
	var vals Values
	n, err := ts.schema.Decode(src, &vals)
	if err != nil {
		return 0, err
	}
	for _, f := range ts.fields {
		f.decode(f.field, &vals, v)
	}
	return n, nil
}
