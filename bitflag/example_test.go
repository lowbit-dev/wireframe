package bitflag_test

import (
	"fmt"

	"lowbit.dev/wireframe/bitflag"
)

// NetFlags represents TCP-like control flags packed in a single byte.
type NetFlags uint8

const (
	FlagSYN NetFlags = 1 << 0
	FlagACK NetFlags = 1 << 1
	FlagFIN NetFlags = 1 << 2
)

func Example() {
	f := bitflag.Of(FlagSYN)
	f.Set(FlagACK)

	fmt.Println(f.Has(FlagSYN))
	fmt.Println(f.Has(FlagFIN))
	f.Clear(FlagSYN)
	fmt.Println(f.Value()) // only FlagACK remains
	// Output:
	// true
	// false
	// 2
}

func ExampleOf() {
	f := bitflag.Of(FlagSYN, FlagACK)
	fmt.Println(f.Has(FlagSYN))
	fmt.Println(f.Has(FlagFIN))
	fmt.Println(f.Value()) // FlagSYN | FlagACK = 3
	// Output:
	// true
	// false
	// 3
}

func ExampleFlags_Has() {
	f := bitflag.Of(FlagSYN)
	fmt.Println(f.Has(FlagSYN))
	fmt.Println(f.Has(FlagACK))
	// Output:
	// true
	// false
}

func ExampleFlags_HasAll() {
	f := bitflag.Of(FlagSYN, FlagACK)
	fmt.Println(f.HasAll(FlagSYN, FlagACK))
	fmt.Println(f.HasAll(FlagSYN, FlagFIN))
	// Output:
	// true
	// false
}

func ExampleFlags_HasAny() {
	f := bitflag.Of(FlagSYN)
	fmt.Println(f.HasAny(FlagSYN, FlagFIN))
	fmt.Println(f.HasAny(FlagACK, FlagFIN))
	// Output:
	// true
	// false
}

func ExampleFlags_Set() {
	var f bitflag.Flags[NetFlags]
	f.Set(FlagSYN)
	f.Set(FlagFIN)
	fmt.Println(f.Value()) // FlagSYN | FlagFIN = 5
	// Output:
	// 5
}

func ExampleFlags_Clear() {
	f := bitflag.Of(FlagSYN, FlagACK, FlagFIN)
	f.Clear(FlagACK)
	fmt.Println(f.Has(FlagACK))
	fmt.Println(f.Value()) // FlagSYN | FlagFIN = 5
	// Output:
	// false
	// 5
}

func ExampleFlags_Value() {
	f := bitflag.Of(FlagSYN, FlagACK)
	fmt.Println(f.Value()) // 1 | 2 = 3
	// Output:
	// 3
}
