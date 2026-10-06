# Conformance corpus

Portable `RICE` modules compiled from `exec/testdata/*.rice`, plus port-rule vectors.

```
go test ./exec/conformance -update
```

rebuilds `testdata/*.ricebc`, `vectors/*.ricebc`, and `expected.txt`. Other ports load the same bytes and match `expected.txt`. `-update` does not touch `compat/`.

## Frozen compatibility

`compat/<major.minor>/` holds blobs and an `expected.txt` from that version. A VM or port at version `M.n` must pass every folder whose version it supports (same rule as `Decode`: older majors down to `MinMajor`, and the same major with a minor not newer than `n`). Each future release adds its own folder; existing folders stay byte-for-byte.
