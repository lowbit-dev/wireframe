package checksum_test

import (
	"fmt"

	"lowbit.dev/wireframe/checksum"
)

func Example() {
	// Algorithms are interchangeable through the Algorithm interface.
	alg := checksum.CRC32IEEE()
	data := []byte("binary payload")
	sum := alg.Append(nil, data)
	fmt.Println(alg.Verify(data, sum))
	// Output:
	// true
}

func ExampleCRC32IEEE() {
	alg := checksum.CRC32IEEE()
	data := []byte("hello")
	sum := alg.Append(nil, data)
	fmt.Println(len(sum))
	fmt.Println(alg.Verify(data, sum))
	// Output:
	// 4
	// true
}

func ExampleCRC64ISO() {
	alg := checksum.CRC64ISO()
	data := []byte("hello")
	sum := alg.Append(nil, data)
	fmt.Println(len(sum))
	fmt.Println(alg.Verify(data, sum))
	// Output:
	// 8
	// true
}

func ExampleXOR8() {
	alg := checksum.XOR8()
	// 0x01 XOR 0x02 = 0x03
	data := []byte{0x01, 0x02}
	sum := alg.Append(nil, data)
	fmt.Printf("%02X\n", sum[0])
	fmt.Println(alg.Verify(data, sum))
	// Output:
	// 03
	// true
}

func ExampleCRC16() {
	// Modbus RTU known vector: device 0x11, FC 0x03, start 0x006B, count 3.
	alg := checksum.CRC16()
	data := []byte{0x11, 0x03, 0x00, 0x6B, 0x00, 0x03}
	sum := alg.Append(nil, data)
	fmt.Printf("%02X %02X\n", sum[0], sum[1])
	// Output:
	// 76 87
}

func ExampleAdler32() {
	alg := checksum.Adler32()
	data := []byte("hello")
	sum := alg.Append(nil, data)
	fmt.Println(len(sum))
	fmt.Println(alg.Verify(data, sum))
	// Output:
	// 4
	// true
}
