package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
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

	var folderMap map[string]*DirNode
	folderMap, err = start_scanning(absRoot, *limit)
	runREPL(folderMap, absRoot)
}
