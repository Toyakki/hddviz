package main

import (
	"container/heap"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

// Declaring data type is a peak coding practice
type ByteSize float64

const (
	_            = iota
	KiB ByteSize = 1 << (10 * iota)
	MiB
	GiB
	TiB
)

// Convert bytes to KiB, MiB, GiB, or TiB for readability
func sizeify(size ByteSize) string {
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
		return fmt.Sprintf("%.2f bytes", size)
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

	// Ensure file is displayed in reverse size order.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Recursive bottom-up approach to compute each directory size
// and its topK largest subdirectories

// Need to avoid symbolic and hard links for double counting.
func scanDir(
	fileSystem fs.FS,
	parentPath string,
	folderMap map[string]*DirNode,
	limit int,
	stats *ScanStats,
) (int64, error) {
	var totalSize int64
	h := &TupleHeap{}
	heap.Init(h)

	entries, err := fs.ReadDir(fileSystem, parentPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			stats.Skipped = append(stats.Skipped, parentPath)
			return 0, fmt.Errorf("skipped permission denied for the path %q:  %w", parentPath, err)
		} else if errors.Is(err, fs.ErrNotExist) {
			stats.Skipped = append(stats.Skipped, parentPath)
			return 0, fmt.Errorf("skipped non-existent directory for the path %q:  %w", parentPath, err)
		}
		return 0, fmt.Errorf("unexpected error for the path %q:  %w", parentPath, err)
	}
	for _, entry := range entries {
		// Avoid symlink just in case
		if entry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		childPath := filepath.Join(parentPath, entry.Name())
		if entry.IsDir() {
			childSize, err := scanDir(
				fileSystem,
				childPath,
				folderMap,
				limit,
				stats,
			)

			if err != nil {
				var pathError *fs.PathError
				if errors.As(err, &pathError) {
					continue
				}
				// Avoid unnecessary wrapping from the recursive call.
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
				if errors.Is(err, fs.ErrPermission) {
					continue
				}
				return 0, fmt.Errorf("unexpected error when opening a file %q: %w", info, err)
			}
			totalSize += info.Size()
		}
	}
	folderMap[parentPath] = &DirNode{
		FolderName:   filepath.Base(parentPath),
		Size:         totalSize,
		TopKChildren: filterChildren(h),
	}

	fmt.Printf("Scanned %s: %s\n", filepath.Base(parentPath), sizeify(ByteSize(totalSize)))
	return totalSize, nil
}

func startScanning(fileSystem fs.FS, absRoot string, limit int) (map[string]*DirNode, error) {
	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}
	if _, err := scanDir(fileSystem, absRoot, folderMap, limit, stats); err != nil {
		// Return the wrapped error
		return nil, err
	}
	if len(stats.Skipped) > 0 {
		fmt.Printf("\nSkipped %d directories/files due to access issues. Consider running 'sudo hddviz' to bypass the permission issue.\n", len(stats.Skipped))
	}
	fmt.Println("")
	fmt.Println("Scanning completed!")
	return folderMap, nil
}
