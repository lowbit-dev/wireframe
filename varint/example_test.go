package varint_test

import (
	"fmt"

	"lowbit.dev/wireframe/varint"
)

func ExampleEncodeU64() {
	buf := make([]byte, 10)
	n := varint.EncodeU64(buf, 300)
	value, _, _ := varint.DecodeU64(buf[:n])
	fmt.Println(n)
	fmt.Println(value)
	// Output:
	// 2
	// 300
}

func ExampleDecodeU64() {
	buf := make([]byte, 10)
	varint.EncodeU64(buf, 16384)
	value, n, _ := varint.DecodeU64(buf)
	fmt.Println(value)
	fmt.Println(n)
	// Output:
	// 16384
	// 3
}

func ExampleEncodeI64() {
	// ZigZag encoding: -1 maps to 1, fitting in a single byte.
	buf := make([]byte, 10)
	n := varint.EncodeI64(buf, -1)
	value, _, _ := varint.DecodeI64(buf[:n])
	fmt.Println(n)
	fmt.Println(value)
	// Output:
	// 1
	// -1
}

func ExampleDecodeI64() {
	buf := make([]byte, 10)
	varint.EncodeI64(buf, -64)
	value, n, _ := varint.DecodeI64(buf)
	fmt.Println(value)
	fmt.Println(n) // -64 ZigZag-encodes to 127, still one byte
	// Output:
	// -64
	// 1
}

func ExampleSizeU64() {
	fmt.Println(varint.SizeU64(0))
	fmt.Println(varint.SizeU64(127))
	fmt.Println(varint.SizeU64(128))
	fmt.Println(varint.SizeU64(16383))
	fmt.Println(varint.SizeU64(16384))
	// Output:
	// 1
	// 1
	// 2
	// 2
	// 3
}

func ExampleSizeI64() {
	// ZigZag: 0→0, -1→1, -64→127 all fit in one byte; 64→128 needs two.
	fmt.Println(varint.SizeI64(0))
	fmt.Println(varint.SizeI64(-1))
	fmt.Println(varint.SizeI64(-64))
	fmt.Println(varint.SizeI64(64))
	// Output:
	// 1
	// 1
	// 1
	// 2
}
