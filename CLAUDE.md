# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go library that provides a mutable, in-memory BSON document model built on top of the official MongoDB Go driver (`go.mongodb.org/mongo-driver`). Unlike the driver's raw `bsoncore.Document` (immutable byte slices), this package lets you read, modify, and serialize BSON documents using `Document` and `Array` types backed by maps of `*Element` pointers.

There is no `go.mod` — this is vendored/embedded as part of a larger module.

## Build & Test

```
go test ./...                         # run all tests
go test -run TestNew                  # run a single test
go test -bench .                      # run benchmarks
go vet ./...                          # static analysis
```

## Architecture

### Two layers

1. **Low-level wire format** (`core.go`) — pure-function `Append*`/`Read*` helpers that operate directly on `[]byte` slices in BSON wire format. Ported from `bsoncore` with a local `Type` alias and additional helpers (`ReadTime`, `AppendTime`).

2. **Mutable document model** (`document.go`, `array.go`, `element.go`, `value.go`) — the main API consumers use.

### Mutable model types

- `Document` (`map[string]*Element`) — BSON document. Supports dot-notation access (`doc.Get("info.vip")`, `doc.Set("info.vip", 100)`).
- `Array` (`map[string]*Element`) — BSON array, keyed by string index ("0", "1", …). Ordered iteration uses `strconv.Itoa(i)` from 0 to `len(arr)`.
- `Element` — wraps a `Type` byte and a value (`any`). For leaf types the value is `[]byte`; for embedded documents/arrays the value is a `Document` or `Array` that is lazily created via `getOrCreateDoc`/`getOrCreateArr`.
- `Value` — simple struct of `Type` + `Data []byte`, used in the low-level layer.

### Key data flow

```
Go struct ──bson.MarshalValue──▶ (Type, []byte) ──Element.Reset──▶ Element
                                                                      │
Document.Raw(nil) ◀── bsoncore wire bytes ◀── serializes tree ◀──────┘
```

- `Marshal(interface{})` converts any Go value to a `Document` by marshaling through the official driver then parsing the raw bytes.
- `Document.Raw(nil)` / `Array.Raw(nil)` re-serializes the mutable tree back to `[]byte` wire format.
- `Document.Unmarshal(val)` round-trips through `Raw` then the driver's `bson.Unmarshal`.

### Dot-notation key splitting

`Split(key)` splits on the first `.` — e.g. `"info.vip"` → `("info", "vip")`. `Get`, `Set`, `Unset`, and `loadOrCreate` recurse through the tree using this.

### marshal.go / unmarshal.go

These are near-complete copies of the MongoDB driver's `bson` package marshal/unmarshal surface (with `MarshalValue` returning the local `Type` alias). They exist so this package can be used as a drop-in replacement where both document-level and value-level marshal/unmarshal are needed without importing `go.mongodb.org/mongo-driver/bson` directly.

### Array growth guard

`ArrayAppendLimit` (default 100) caps how many null-fill elements `Array.append` will create when setting a sparse index, preventing accidental huge allocations.
