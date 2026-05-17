# Examples

Runnable examples showing how to build a binary protocol with `wireframe`.

| Example                             | API     | Description                                                               |
| ----------------------------------- | ------- | ------------------------------------------------------------------------- |
| [msgproto](msgproto/)               | Typed   | Full protocol using `TypedFormat[T]`, `TypedSchema[T]`, and `TypedReader` |
| [msgproto_values](msgproto_values/) | Untyped | Same protocol using `frame.Format`, `schema.Values`, and `stream.Reader`  |

Both examples implement the same wire format and produce identical bytes. The only difference is how the Go side interacts with the codec.

## Wire layout

```text
AA 55 | version(u8) | type(u8) | flags(u8) | length(u32) | payload | CRC32
```

## Which example to read first

Start with `msgproto` if you want to see the full stack with the least boilerplate. The typed API is the common path for protocols with a fixed layout.

Read `msgproto_values` after if you need dynamic formats, runtime-configurable field sets, or want to understand what the typed layer is doing underneath.
