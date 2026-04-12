package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := flag.String("root", home, "root directory to scan")
	limit := flag.Int("limit", 10, "number of largest subfolders to keep per directory")

	flag.Parse()
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid abs path: ", err)
		os.Exit(1)
	}

	folderMap, err := start_scanning(absRoot, *limit)
	if err != nil {
		cause := errors.Unwrap(err)
		fmt.Fprintln(os.Stderr, "Scanning failed: ", cause)
		os.Exit(1)
	}
	runREPL(folderMap, absRoot)
}
