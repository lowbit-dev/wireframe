package frame_test

import (
	"fmt"

	"lowbit.dev/wireframe/frame"
)

func Example() {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 65535,
	}

	var vals frame.Values
	frame.Set(&vals, "type", uint8(1))
	encoded, _ := f.Encode(nil, []byte("hello"), &vals)

	payload, dvals, _, _ := f.Decode(encoded)
	fmt.Println(string(payload))
	fmt.Println(dvals.Uint8("type"))
	// Output:
	// hello
	// 1
}

// ExampleValues demonstrates zero-allocation usage: declare on the stack,
// populate with Set, read back typed values, and Reset for reuse.
func ExampleValues() {
	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(0x42))

	fmt.Println(vals.Uint8("version"))
	fmt.Println(vals.Uint8("type"))

	vals.Reset()
	fmt.Println(vals.Uint8("version")) // zero after reset
	// Output:
	// 1
	// 66
	// 0
}

func ExampleSet() {
	var vals frame.Values
	frame.Set(&vals, "flags", uint8(0b00000101))
	fmt.Printf("%08b\n", vals.Uint8("flags"))
	// Output:
	// 00000101
}

func ExampleFormat_Encode() {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 65535,
	}

	var vals frame.Values
	frame.Set(&vals, "type", uint8(7))
	encoded, err := f.Encode(nil, []byte("hello"), &vals)
	fmt.Println(err)
	fmt.Println(len(encoded) > 0)
	// Output:
	// <nil>
	// true
}

func ExampleFormat_Decode() {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 65535,
	}

	var vals frame.Values
	frame.Set(&vals, "type", uint8(7))
	encoded, _ := f.Encode(nil, []byte("hello"), &vals)

	payload, dvals, _, err := f.Decode(encoded)
	fmt.Println(err)
	fmt.Println(string(payload))
	fmt.Println(dvals.Uint8("type"))
	// Output:
	// <nil>
	// hello
	// 7
}

// ExampleFormat_DecodeInto demonstrates the zero-allocation decode path
// used by stream.Reader: a caller-supplied Values is reset and reused.
func ExampleFormat_DecodeInto() {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 65535,
	}

	var vals frame.Values
	frame.Set(&vals, "type", uint8(3))
	encoded, _ := f.Encode(nil, []byte("ping"), &vals)

	var reusable frame.Values
	payload, _, err := f.DecodeInto(encoded, &reusable)
	fmt.Println(err)
	fmt.Println(string(payload))
	fmt.Println(reusable.Uint8("type"))
	// Output:
	// <nil>
	// ping
	// 3
}
