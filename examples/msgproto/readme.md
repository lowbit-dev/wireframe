# msgproto

A complete binary messaging protocol built with `wireframe`, using the typed API throughout.

Wire layout:

```text
AA 55 | version(u8) | type(u8) | flags(u8) | length(u32) | payload | CRC32
```

Three message types are defined:

| Type    | Value | Payload                       |
| ------- | ----- | ----------------------------- |
| Ping    | 0x01  | none; FlagPriority may be set |
| Data    | 0x02  | seq(u16BE) + data(var bytes)  |
| DataAck | 0x03  | seq(u16BE)                    |

## What this example covers

- `frame.NewTypedFormat[Header]` — declare the frame envelope once; encode and decode using a plain struct, no string keys
- `schema.NewTyped[T]` — describe the payload layout with struct-bound fields; the compiler checks field types at the definition site
- `stream.NewTypedReader` — read frames off a connection incrementally; each `Next()` returns a fully populated header struct by value
- `bitflag.Of` — set and test named flag bits on the header flags field without raw bit arithmetic

## Run

```sh
go run lowbit.dev/wireframe/examples/msgproto
```

From the module root:

```sh
go run ./examples/msgproto/
```

Expected output:

```
encoded 117 bytes into buffer

PING    version=1 priority=true
DATA    seq=1 data="hello from message 1"
DATA    seq=2 data="hello from message 2"
ACK     seq=1
ACK     seq=2
```

## Compare

See `examples/msgproto_values` for the same protocol implemented with the untyped `Values`-based API. Both produce identical wire bytes.
