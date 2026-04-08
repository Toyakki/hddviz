// This is a crazy project since I am using a lc data structure :)

package main

import (
	"bufio"
	"container/heap"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Constants, easy to configure.
const KiB = 1 << 10
const MiB = 1 << 20
const GiB = 1 << 30
const TiB = 1 << 40

type Config struct {
	ExcludedFolders []string `json:"excluded_folders"`
}

func readConfig(filename string) (*Config, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := jsonFile.Close(); err != nil {
		fmt.Println("Error closing file:", err)
		return nil, err
	}
	byteVals, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	var config Config
	err = json.Unmarshal(byteVals, &config)
	if err != nil {
		fmt.Println("JSON parsing failed: ", err)
		return nil, err
	}
	return &config, nil
}

func constructFolderPaths(configFileName string, root string) (map[string]struct{}, error) {
	excludeSet := make(map[string]struct{})
	config, err := readConfig(configFileName)
	if err != nil {
		return nil, err
	}
	for _, folder := range config.ExcludedFolders {
		folderPath := filepath.Join(root, folder)
		excludeSet[folderPath] = struct{}{}
	}
	return excludeSet, nil
}

// Convert bytes to KiB, MiB, GiB, or TiB for readability
func sizeify(size int64) string {
	switch {
	case size >= TiB:
		return fmt.Sprintf("%.2f TiB", float64(size)/float64(TiB))
	case size >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(size)/float64(GiB))
	case size >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(size)/float64(MiB))
	case size >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(size)/float64(KiB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func resolvePath(cwd, input string) string {
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	return filepath.Clean(filepath.Join(cwd, input))
}

func printNode(path string, folderMap map[string]*DirNode) {
	node := folderMap[path]
	fmt.Println("path:", path)
	fmt.Println("name:", node.FolderName)
	fmt.Println("size:", sizeify(node.Size))
	fmt.Println("top children:")
	for _, child := range node.TopKChildren {
		fmt.Println(" -", child, "(", sizeify(folderMap[child].Size), ")")
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("help: Show this help")
	fmt.Println("pwd: Show current directory in REPL")
	fmt.Println("ls: Show current directory summary")
	fmt.Println("inspect <path> Show summary for a path")
	fmt.Println("cd <path> Change REPL current directory")
	fmt.Println("quit | exit Leave interactive mode")
}

func runREPL(folderMap map[string]*DirNode, cwd string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println()
	fmt.Println("Entering interactive mode. Type 'help' for commands.")

	for {
		fmt.Printf("hddviz> ")
		if !scanner.Scan() {
			fmt.Println()
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		cmd := parts[0]
		switch cmd {
		case "help":
			printHelp()

		case "quit", "exit":
			return
		case "pwd":
			fmt.Println(cwd)
		case "ls":
			printNode(cwd, folderMap)
		case "inspect":
			if len(parts) < 2 {
				fmt.Println("usage: inspect <SUBFOLDER_NAME>")
				continue
			}
			target := resolvePath(cwd, parts[1])
			if _, ok := folderMap[target]; !ok {
				fmt.Println("directory not found in scan: ", target)
				continue
			}
			printNode(target, folderMap)
		case "cd":
			if len(parts) < 2 {
				fmt.Println("usage: cd <abs path >")
				continue
			}
			target := resolvePath(cwd, parts[1])
			if _, ok := folderMap[target]; !ok {
				fmt.Println("directory not found in scan: ", target)
				continue
			}
			cwd = target
		default:
			fmt.Println("unknown command: ", cmd)
		}
	}
}

type DirNode struct {
	FolderName   string
	Size         int64
	TopKChildren []string
}

type ScanStats struct {
	Skipped []string
}

// DFS bottom-up approach to compute the directory size and its topk largest subdirectories.
func scanDir(
	parentPath string,
	folderMap map[string]*DirNode,
	excludeSet map[string]struct{},
	limit int,
	stats *ScanStats,
) (int64, error) {
	var totalSize int64
	h := &TupleHeap{}
	heap.Init(h)

	entries, err := os.ReadDir(parentPath)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			stats.Skipped = append(stats.Skipped, parentPath)
			fmt.Println("Skipped permission denied:", parentPath)
			return 0, nil
		}
		return 0, err
	}
	for _, entry := range entries {
		childPath := filepath.Join(parentPath, entry.Name())
		if entry.IsDir() {
			// Check if the directory is in the exclude list.
			if _, excluded := excludeSet[childPath]; excluded {
				continue
			}
			childSize, err := scanDir(
				childPath,
				folderMap,
				excludeSet,
				limit,
				stats,
			)
			if err != nil {
				return 0, err
			}
			totalSize += childSize
			heap.Push(h, ChildTuple{Path: childPath, Size: childSize})
			if h.Len() > limit {
				heap.Pop(h)
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					continue
				}
				return 0, err
			}
			totalSize += info.Size()
		}
	}
	folderMap[parentPath] = &DirNode{
		FolderName:   filepath.Base(parentPath),
		Size:         totalSize,
		TopKChildren: filterChildren(h),
	}
	fmt.Printf("Scanned %s: %s\n", filepath.Base(parentPath), sizeify(totalSize))
	return totalSize, nil
}

type ChildTuple struct {
	Path string
	Size int64
}

type TupleHeap []ChildTuple

func (h TupleHeap) Len() int { return len(h) }

// Make sure that each subdirectory's size is compared.
func (h TupleHeap) Less(i, j int) bool { return h[i].Size < h[j].Size }
func (h TupleHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TupleHeap) Push(x any) {
	// Retrieve the underlying concrete value from an interface variable
	*h = append(*h, x.(ChildTuple))
}

func (h *TupleHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func filterChildren(h *TupleHeap) []string {
	n := h.Len()
	out := make([]string, n)
	for i := 0; h.Len() > 0; i++ {
		out[i] = heap.Pop(h).(ChildTuple).Path
	}
	// Reverse order largest -> smallest.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func main() {
	// Home directory creation
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	root := flag.String("root", home, "root directory to scan")
	limit := flag.Int("limit", 10, "number of largest subfolders to keep per directory")
	config := flag.String("config", "config.json", "path to exclude config")
	once := flag.Bool("once", false, "scan and exit without entering interactive mode")

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

	excludeSet, err := constructFolderPaths(*config, absRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
	}
	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}
	if _, err := scanDir(absRoot, folderMap, excludeSet, *limit, stats); err != nil {
		fmt.Fprintln(os.Stderr, "scan failed:", err)
		os.Exit(1)
	}
	if len(stats.Skipped) > 0 {
		fmt.Printf("\nSkipped %d directories due to access issues.\n", len(stats.Skipped))
	}
	fmt.Printf("")
	fmt.Println("Scanning completed!")
	if !*once {
		runREPL(folderMap, absRoot)
	}
}
