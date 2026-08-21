<!-- Documents: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S STK-REQ-260820-T8AZ -->
# Distribution

Phase 1 ships two binaries.

## Native Go binary

```
go build -o crm ./cmd/crm
./crm init
./crm serve
```

This is the supported development path. `crm init`, `crm serve`, and the rest of the CLI work from this binary.

## Cosmopolitan APE (`crm.com`)

A true cosmocc compile of the Go runtime is not available. `make ape` builds a real APE wrapper with Cosmopolitan libc and embeds the host Go binary as a zip member.

```
# Install cosmocc once
mkdir -p "$HOME/.local/cosmocc"
curl -fsSL https://cosmo.zip/pub/cosmocc/cosmocc.zip -o /tmp/cosmocc.zip
unzip -q /tmp/cosmocc.zip -d "$HOME/.local/cosmocc"
chmod -R +x "$HOME/.local/cosmocc/bin" "$HOME/.local/cosmocc/libexec"
export PATH="$HOME/.local/cosmocc/bin:$HOME/.local/bin:$PATH"

# Apple Silicon needs the ape loader
mkdir -p "$HOME/.local/bin"
cc -O -o "$HOME/.local/bin/ape" "$HOME/.local/cosmocc/bin/ape-m1.c"

make ape
chmod +x crm.com
./crm.com --help
```

`scripts/build-ape.sh` also writes `.build/ape/hello.com` so the toolchain can be checked independently.

### What works now

- This Mac (darwin/arm64): `curl .../crm.com && chmod +x crm.com && ./crm.com` extracts the embedded darwin-arm64 blob and execs it.
- `hello.com` is a portable APE built only with cosmocc.
- KI-1 is closed: APE packaging exists and runs on darwin/arm64.

### Remaining OS targets

The wrapper looks for `/zip/crm.<os>-<arch>` then `/zip/crm`. Cross-OS blobs are not embedded yet:

- Linux x86-64 / ARM64
- macOS x86-64
- Windows x86-64
- supported BSD targets

Add those blobs to the zip overlay to extend host coverage. Windows ARM64 is not a guaranteed Cosmopolitan target.
