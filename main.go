// This is a crazy project since I am using a lc data structure :)

package main

import (
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Constants, easy to configure.
const KiB = 1 << 10
const MiB = 1 << 20
const GiB = 1 << 30
const TiB = 1 << 40

const configFileName = "excluded_folders.yaml"

// Number of folders to show for each circle chart. You can increase this param.
const viz_limit = 10

// Showing it is scanning or not.
const verbose = true

type HiddenFolderNames struct {
	Mac struct {
		Include []string `yaml:"include"`
		Exclude []string `yaml:"exclude"`
	} `yaml:"mac"`
}

func readHiddenFolderNames(filename string) (*HiddenFolderNames, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var names HiddenFolderNames

	if err := yaml.Unmarshal(buf, &names); err != nil {
		return nil, err
	}
	return &names, nil
}

func constructFolderPaths(configFileName string, root string) map[string]struct{} {
	folders, err := readHiddenFolderNames(configFileName)
	if err != nil {
		panic(err)
	}
	excludeSet := make(map[string]struct{}, len(folders.Mac.Exclude))
	for _, folder := range folders.Mac.Exclude {
		folderPath := filepath.Join(root, folder)
		// Empty struct literal to occupy zero bytes and to flag as True in exclusion search.
		excludeSet[folderPath] = struct{}{}
	}
	return excludeSet
}

type DirNode struct {
	FolderName   string
	Size         int64
	TopKChildren []string
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

// DFS bottom-up approach to compute the directory size and its topk largest subdirectories.
func scanDir(
	parentPath string,
	folderMap map[string]*DirNode,
	excludeSet map[string]struct{},
	limit int,
	verbose bool) (int64, error) {
	var totalSize int64
	h := &TupleHeap{}
	heap.Init(h)

	entries, err := os.ReadDir(parentPath)
	if err != nil {
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
				verbose,
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
	if verbose {
		fmt.Printf("Scanned %s: %s\n", filepath.Base(parentPath), sizeify(totalSize))
	}
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
	out := make([]string, h.Len())
	for i := 0; h.Len() > 0; i++ {
		out[i] = heap.Pop(h).(ChildTuple).Path
	}
	return out
}

func generateJson(data any, fileName string) error {
	outputFilePath := filepath.Join("../frontend", fileName)
	f, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)

	// For production, no setindent
	encoder.SetIndent("", " ")
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
	return nil
}

func main() {
	// Allow user's input to specify their username.
	var username string
	fmt.Print("Enter your username: ")
	fmt.Scanln(&username)
	// Basic sanitization
	username = strings.TrimSpace(username)

	// Home directory creation
	root := filepath.Join("/Users", username)
	_, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Cannot find your home directory with the specified username. Did you type it correctly?")
			os.Exit(1)
		}
	}
	excludeSet := constructFolderPaths(configFileName, root)
	folderMap := make(map[string]*DirNode)
	if _, err := scanDir(root, folderMap, excludeSet, viz_limit, verbose); err != nil {
		panic(err)
	}
	if err := generateJson(folderMap, "foldermap.json"); err != nil {
		panic(err)
	}

	// fix the deadlock!
	// parallelism := max(4, runtime.GOMAXPROCS(0)*2)
	// sem := make(chan struct{}, parallelism)
	// var mu sync.Mutex

	// if _, err := scanDirConcurrent(root, folderMap, excludeSet, viz_limit, verbose, sem, &mu); err != nil {
	// 	panic(err)
	// }

	// if err := generateJson(folderMap, "foldermap.json"); err != nil {
	// 	panic(err)
	// }
}
