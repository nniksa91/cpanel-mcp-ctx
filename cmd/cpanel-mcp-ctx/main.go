package main

import (
	"fmt"
	"os"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/cli"
)

var version = "0.1.0-dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
