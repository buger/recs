<!-- Documents: STK-REQ-260820-T8AZ -->
# Distribution

Phase 1 ships one native Go binary.

```
go build -o recs ./cmd/recs
./recs init
./recs serve
```

This is the only supported build path. Cosmopolitan, APE, cosmocc, and `recs.com` are permanent non-goals.
