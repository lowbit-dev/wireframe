package frame

// maxInlineFields is the capacity of the inline key/value arrays inside
// Values. Protocol headers are almost never larger than this; the inline
// path avoids all heap allocations for the common case.
const maxInlineFields = 16

type rawValue struct {
	u64 uint64
	b   []byte
}

// Values carries header field values between the caller and the codec.
//
// During encoding the caller populates it before calling Format.Encode.
// During decoding the codec populates it from the wire bytes.
//
// Values is a value type. Declare it on the stack and call Reset between
// uses to avoid heap allocations. Use NewValues when a heap-allocated
// instance is genuinely needed (e.g. storing across goroutine boundaries).
type Values struct {
	keys [maxInlineFields]string
	vals [maxInlineFields]rawValue
	n    int
	// overflow handles the rare case of more than maxInlineFields fields.
	overflow map[string]rawValue
}

// NewValues allocates a new Values on the heap.
func NewValues() *Values {
	return &Values{}
}

// Reset clears all stored values, making Values ready for reuse.
func (v *Values) Reset() {
	v.n = 0
	v.overflow = nil
}

// set stores a raw value for key, replacing any existing value.
func (v *Values) set(key string, rv rawValue) {
	// Update existing inline entry.
	for i := 0; i < v.n; i++ {
		if v.keys[i] == key {
			v.vals[i] = rv
			return
		}
	}
	// Update existing overflow entry.
	if v.overflow != nil {
		if _, ok := v.overflow[key]; ok {
			v.overflow[key] = rv
			return
		}
	}
	// Insert new entry.
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

// get retrieves the raw value for key.
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

// Set stores v under key in vals. Using a package-level generic function
// rather than a method is required because Go does not allow type
// parameters on methods of a non-generic type. Type mismatches are caught
// at compile time.
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
	default:
		// User-defined types with unsigned integer bases reach this branch.
		// Use a safe runtime conversion via the constraint.
		setUnsigned(vals, key, v)
		return
	}
	vals.set(key, rv)
}

// setUnsigned handles user-defined types that satisfy Settable but do not
// match one of the built-in cases in the type switch above.
func setUnsigned[T Settable](vals *Values, key string, v T) {
	// Convert through uint64 — valid for all unsigned integer base types.
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
