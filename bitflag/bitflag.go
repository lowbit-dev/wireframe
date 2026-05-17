// Package bitflag provides a generic, type-safe bitset for flag fields
// in binary protocol headers.
//
// Flags are parameterised over any unsigned integer type, including
// user-defined named types. This lets the compiler reject accidental
// mixing of flag constants from different fields.
//
//	type MsgFlags uint8
//
//	const (
//	    FlagAck      MsgFlags = 1 << 0
//	    FlagRetry    MsgFlags = 1 << 1
//	    FlagCompress MsgFlags = 1 << 2
//	)
//
//	f := bitflag.Of(FlagAck, FlagCompress)
//	f.Has(FlagAck)      // true
//	f.Has(FlagRetry)    // false
//	raw := f.Value()    // MsgFlags(0b00000101)
package bitflag

// Integer is the constraint satisfied by unsigned integer types and
// user-defined types with an unsigned integer base.
type Integer interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Flags is a type-safe bitset backed by an unsigned integer of type T.
type Flags[T Integer] struct {
	v T
}

// Of constructs a Flags value with the given flags set.
func Of[T Integer](flags ...T) Flags[T] {
	var f Flags[T]
	f.Set(flags...)
	return f
}

// Has reports whether flag is set. Exactly one bit should be set in flag.
func (f Flags[T]) Has(flag T) bool {
	return f.v&flag != 0
}

// HasAll reports whether all of the given flags are set.
func (f Flags[T]) HasAll(flags ...T) bool {
	for _, flag := range flags {
		if f.v&flag == 0 {
			return false
		}
	}
	return true
}

// HasAny reports whether at least one of the given flags is set.
func (f Flags[T]) HasAny(flags ...T) bool {
	for _, flag := range flags {
		if f.v&flag != 0 {
			return true
		}
	}
	return false
}

// Set sets the given flags, leaving others unchanged.
func (f *Flags[T]) Set(flags ...T) {
	for _, flag := range flags {
		f.v |= flag
	}
}

// Clear clears the given flags, leaving others unchanged.
func (f *Flags[T]) Clear(flags ...T) {
	for _, flag := range flags {
		f.v &^= flag
	}
}

// Value returns the underlying integer representation.
func (f Flags[T]) Value() T {
	return f.v
}
