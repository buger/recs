package dist

// BuildAPE documents the Cosmopolitan wrapper path.
// The real build lives in scripts/build-ape.sh and ape/wrapper.c.
// Implements: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S
func BuildAPE() string {
	return "scripts/build-ape.sh"
}
