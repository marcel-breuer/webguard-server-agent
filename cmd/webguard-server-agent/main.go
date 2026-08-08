package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(version)
		return
	}

	fmt.Fprintln(os.Stderr, "WebGuard Server Agent is not configured yet. Run with --version to inspect the build version.")
	os.Exit(2)
}
