# msgproto_values

The same binary messaging protocol as `examples/msgproto`, implemented with the untyped `Values`-based API. Both examples produce identical wire bytes.

Wire layout:

```text
AA 55 | version(u8) | type(u8) | flags(u8) | length(u32) | payload | CRC32
```

## What this example covers

- `frame.Format` — declare the frame envelope as a plain struct literal; field values pass through `frame.Values` using string keys
- `schema.New` — describe the payload layout with named fields; values are read and written via `schema.Values`
- `stream.NewReader` — read frames incrementally; `Next()` returns a pointer to the reader's internal `Values`, valid until the next call
- `bitflag.Of` — same flag handling as the typed example; `bitflag` is independent of which codec API you use

## When to use this API

The `Values`-based API is the right choice when:

- The format is not fully known at compile time (e.g. loaded from config)
- Fields vary per connection or message version
- You want to keep format declarations and application structs separate
- You are working at a layer that dispatches to multiple typed handlers and only needs a few header fields

For formats with a fixed, well-known layout the typed API in `examples/msgproto` is less code and catches field mismatches at compile time.

## Run

```sh
go run lowbit.dev/wireframe/examples/msgproto_values
```

From the module root:

```sh
go run ./examples/msgproto_values/
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
