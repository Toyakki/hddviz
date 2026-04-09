package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Home directory creation
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	root := flag.String("root", home, "root directory to scan")
	limit := flag.Int("limit", 10, "number of largest subfolders to keep per directory")

	flag.Parse()
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid root: ", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "root does not exist: ", absRoot)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "cannot access root: ", err)
		os.Exit(1)
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}
	if _, err := scanDir(absRoot, folderMap, *limit, stats); err != nil {
		fmt.Fprintln(os.Stderr, "scan failed:", err)
		os.Exit(1)
	}
	if len(stats.Skipped) > 0 {
		fmt.Printf("\nSkipped %d directories due to access issues.\n", len(stats.Skipped))
	}
	fmt.Printf("")
	fmt.Println("Scanning completed!")
	runREPL(folderMap, absRoot)
}
