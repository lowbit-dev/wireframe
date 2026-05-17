package schema_test

import (
	"fmt"

	"lowbit.dev/wireframe/schema"
)

func Example() {
	s := schema.New(
		schema.Uint8("op"),
		schema.Uint16BE("id"),
	)

	vals := schema.NewValues()
	schema.Set(vals, "op", uint8(1))
	schema.Set(vals, "id", uint16(42))
	encoded, _ := s.Encode(nil, vals)

	decoded := schema.NewValues()
	s.Decode(encoded, decoded)
	fmt.Println(decoded.Uint8("op"))
	fmt.Println(decoded.Uint16("id"))
	// Output:
	// 1
	// 42
}

func ExampleNew() {
	s := schema.New(
		schema.Uint8("command"),
		schema.Uint16BE("seq"),
	)

	vals := schema.NewValues()
	schema.Set(vals, "command", uint8(3))
	schema.Set(vals, "seq", uint16(42))
	encoded, _ := s.Encode(nil, vals)

	decoded := schema.NewValues()
	s.Decode(encoded, decoded)
	fmt.Println(decoded.Uint8("command"))
	fmt.Println(decoded.Uint16("seq"))
	// Output:
	// 3
	// 42
}

func ExampleSchema_Encode() {
	s := schema.New(
		schema.Uint32BE("id"),
		schema.Bool("active"),
	)

	vals := schema.NewValues()
	schema.Set(vals, "id", uint32(1000))
	schema.Set(vals, "active", true)
	encoded, _ := s.Encode(nil, vals)
	fmt.Println(len(encoded)) // 4 bytes (Uint32BE) + 1 byte (Bool)
	// Output:
	// 5
}

func ExampleVarBytes() {
	s := schema.New(
		schema.Uint16BE("len"),
		schema.VarBytes("body", "len"),
	)

	// The length field is populated automatically from the body during encoding.
	vals := schema.NewValues()
	schema.Set(vals, "body", []byte("world"))
	encoded, _ := s.Encode(nil, vals)

	decoded := schema.NewValues()
	s.Decode(encoded, decoded)
	fmt.Println(string(decoded.Bytes("body")))
	// Output:
	// world
}

func ExampleVarint() {
	s := schema.New(schema.Varint("seq"))

	vals := schema.NewValues()
	schema.Set(vals, "seq", uint32(300)) // 300 encodes as 2 bytes
	encoded, _ := s.Encode(nil, vals)
	fmt.Println(len(encoded))

	decoded := schema.NewValues()
	s.Decode(encoded, decoded)
	fmt.Println(decoded.Uint32("seq"))
	// Output:
	// 2
	// 300
}
