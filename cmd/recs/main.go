package main

import (
	"os"

	"github.com/buger/recs/internal/cli"
)

// Implements: INT-REQ-260820-JC9M
func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
