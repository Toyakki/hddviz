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
		fmt.Fprintln(os.Stderr, "Failed to normalize the given path: ", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absRoot); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot find the path: ", err)
		os.Exit(1)
	}
	fsys := os.DirFS(absRoot)

	folderMap, err := startScanning(fsys, *limit)
	if err != nil {
		cause := errors.Unwrap(err)
		fmt.Fprintln(os.Stderr, "Scanning failed: ", cause)
		os.Exit(1)
	}
	runREPL(folderMap, ".")
}
