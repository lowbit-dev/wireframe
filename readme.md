# wireframe

[![Go Report Card](https://goreportcard.com/badge/lowbit.dev/wireframe)](https://goreportcard.com/report/lowbit.dev/wireframe) [![Go Reference](https://pkg.go.dev/badge/lowbit.dev/wireframe.svg)](https://pkg.go.dev/lowbit.dev/wireframe) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A primitive toolkit for building binary protocols in Go.

`wireframe` provides the building blocks that every binary protocol ends up reinventing: frame encoding and decoding, variable-length integers, endian-safe primitive encoding, declarative payload schemas, incremental stream parsers, and checksum validation.

No external dependencies. No reflection-based serialization. No global registries. No code generation.

---

## Packages

```text
wireframe/
├── varint/    variable-length integer encoding
├── endian/    type-safe fixed-width integer helpers
├── bitflag/   generic type-safe bitsets for protocol flag fields
├── checksum/  unified checksum interface with built-in algorithms
├── frame/     frame encoding and decoding (includes TypedFormat[T])
├── stream/    incremental I/O over net.Conn or io.Reader
└── schema/    binary payload layout (includes TypedSchema[T])
```

Each subpackage can be used independently.

---

## Install

```sh
go get lowbit.dev/wireframe
```

Requires Go 1.23+.

---

## Quick start

Define a format, encode a frame, read frames from a connection:

```go
format := &frame.Format{
    Prefix: []byte{0xAA, 0x55},
    Header: []frame.Field{
        frame.Uint8("version"),
        frame.Uint8("type"),
        frame.Uint8("flags"),
        frame.PayloadLength("length"),
    },
    Checksum:   checksum.CRC32IEEE(),
    MaxPayload: 64 * 1024,
}
```

This produces frames of the form:

```text
AA 55 | version | type | flags | length | payload | crc32
```

**Encode:**

```go
var vals frame.Values
frame.Set(&vals, "version", uint8(1))
frame.Set(&vals, "type",    uint8(0x42))
frame.Set(&vals, "flags",   uint8(0))

encoded, err := format.Encode(nil, payload, &vals)
```

**Read from a connection:**

```go
sr := stream.NewReader(conn, format)

for {
    payload, vals, err := sr.Next()
    if err != nil {
        return err
    }
    msgType := vals.Uint8("type")
    handle(msgType, payload)
}
```

---

## Typed API

The `Values`-based API is explicit and allocation-free, but field names are strings. `TypedFormat[T]` and `TypedSchema[T]` bind the codec directly to a struct, catching mismatches at compile time.

```go
type Header struct {
    Version uint8
    Type    uint8
    Flags   uint8
    Length  uint16
}

var msgFormat = frame.NewTypedFormat[Header](
    frame.WithPrefix([]byte{0xAA, 0x55}),
    frame.Uint8Field("version", func(h *Header) *uint8  { return &h.Version }),
    frame.Uint8Field("type",    func(h *Header) *uint8  { return &h.Type }),
    frame.Uint8Field("flags",   func(h *Header) *uint8  { return &h.Flags }),
    frame.PayloadLengthField("length", func(h *Header) *uint16 { return &h.Length }),
    frame.WithChecksum(checksum.CRC32IEEE()),
    frame.WithMaxPayload(64*1024),
)

// Encode
hdr := Header{Version: 1, Type: 0x42}
encoded, err := msgFormat.Encode(nil, payload, &hdr)

// Decode
var hdr Header
payload, n, err := msgFormat.Decode(src, &hdr)
```

Typed stream reading:

```go
sr := stream.NewTypedReader(conn, msgFormat)

for {
    data, hdr, err := sr.Next()
    if err != nil {
        return err
    }
    switch hdr.Type {
    case 0x42:
        handle(hdr, data)
    }
}
```

---

## Subpackages

### `frame`

Core framing codec. Encodes and decodes frames according to a user-defined `Format`. A frame is composed of an optional prefix, zero or more typed header fields, a payload, an optional checksum, and an optional suffix. `frame` handles packet boundaries and integrity — not payload structure.

Built-in fields: `Uint8`, `Uint16BE`, `Uint32BE`, `Uint64BE`, `PayloadLength`.

`Values` is stack-allocatable and reusable. For the common typed path, no per-frame heap allocation occurs.

### `stream`

Incremental frame parsing over an `io.Reader`. Handles buffering and partial reads transparently. `Next` blocks until a complete frame is available. The internal buffer and `Values` are reused across calls — zero per-frame heap allocations for the lifetime of a connection.

### `schema`

Binary payload layout. Describes the structured data inside a frame payload. Independent from `frame`; the two compose naturally.

Built-in field types:

| Constructor                   | Description                         |
| ----------------------------- | ----------------------------------- |
| `Uint8(name)`                 | 1-byte unsigned integer             |
| `Uint16BE(name)`              | 2-byte unsigned integer, big-endian |
| `Uint32BE(name)`              | 4-byte unsigned integer, big-endian |
| `Bool(name)`                  | 1-byte boolean                      |
| `Varint(name)`                | Variable-length integer             |
| `Bytes(name, length)`         | Fixed-length byte slice             |
| `VarBytes(name, lengthField)` | Length-prefixed byte slice          |

### `varint`

Variable-length integer encoding. Wraps `encoding/binary` with a consistent API and adds signed integer support via ZigZag encoding. No allocations.

```go
n := varint.EncodeU64(dst, 12345)
v, n, err := varint.DecodeU64(src)

n := varint.EncodeI64(dst, -42)
v, n, err := varint.DecodeI64(src)
```

### `endian`

Type-safe helpers for fixed-width integer encoding. Works with user-defined types derived from unsigned integers.

```go
endian.PutBE(buf, uint32(0xDEADBEEF))
endian.PutBE(buf, SessionID(42)) // user-defined type, works automatically

v := endian.ReadBE[uint32](buf)
```

### `bitflag`

Generic type-safe bitset for protocol flag fields. Flag constants are checked at compile time — you cannot mix flag types from different fields.

```go
type MsgFlags uint8

const (
    FlagAck      MsgFlags = 1 << 0
    FlagRetry    MsgFlags = 1 << 1
    FlagCompress MsgFlags = 1 << 2
)

f := bitflag.Of(FlagAck, FlagCompress)
f.Has(FlagAck)                  // true
f.HasAll(FlagAck, FlagCompress) // true
f.Set(FlagRetry)
f.Clear(FlagAck)
raw := f.Value()                // MsgFlags(0b00000110)
```

`bitflag` and `frame` are independent. The frame header carries the raw integer; `bitflag` interprets it on either side of the wire.

### `checksum`

Unified interface for checksum algorithms. Built-in implementations:

| Constructor   | Algorithm   |
| ------------- | ----------- |
| `CRC32IEEE()` | CRC-32/IEEE |
| `CRC64ISO()`  | CRC-64/ISO  |
| `XOR8()`      | XOR-8       |
| `CRC16()`     | CRC-16      |
| `Adler32()`   | Adler-32    |

Checksum selection is a configuration decision, not a code change.

---

## Protocol stack

```text
Application Data
        ↓
schema.Schema     — payload structure
        ↓
frame.Format      — framing: header fields, length, checksum
        ↓
stream.Reader     — incremental I/O
        ↓
net.Conn / serial port / file
```

Protocols expressible with this stack include MQTT fixed-header framing, DNS over TCP, Modbus RTU, and custom embedded device protocols.

---

## Errors

All errors are package-level variables, comparable with `errors.Is`.

| Error              | Meaning                      |
| ------------------ | ---------------------------- |
| `ErrIncomplete`    | Not enough bytes to decode   |
| `ErrChecksum`      | Checksum verification failed |
| `ErrFrameTooLarge` | Payload exceeds `MaxPayload` |
| `ErrInvalidData`   | Malformed data               |
| `ErrOverflow`      | Value out of range           |
