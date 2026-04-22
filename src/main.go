package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func printSkipSummary(stats *ScanStats) {
	if stats == nil {
		return
	}
	p := len(stats.PermissionSkip)
	n := len(stats.NoDirSkip)
	f := len(stats.FileInfoSkip)
	if p+n+f == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "\nSkipped %d path(s) (permission denied: %d, not found: %d, file info failed: %d).\n", p+n+f, p, n, f)
	fmt.Fprintln(os.Stderr, "If you expected these to be scanned, re-run with appropriate permissions (sudo hddviz ) or choose a different -root.")
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := flag.String("root", home, "root directory to scan")
	limit := flag.Int("limit", 10, "number of largest subfolders to keep per directory")
	estimate := flag.Bool("est", false, "turn on an estimate mode.")

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

	var folderMap map[string]*DirNode
	var stats *ScanStats

	if *estimate {
		var errs []error
		folderMap, stats, errs = start_scanning_concurrent(fsys, *limit)
		if len(errs) != 0 {
			for _, err := range errs {
				fmt.Fprintln(os.Stderr, "Concurrent scannig failed:", err)
			}
		}
	} else {
		folderMap, stats, err = start_scanning(fsys, *limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Scanning failed: ", err)
			printSkipSummary(stats)
			os.Exit(1)
		}
	}
	printSkipSummary(stats)
	runREPL(folderMap, ".")
}
