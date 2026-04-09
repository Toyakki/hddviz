package main

import (
	"container/heap"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Constants, easy to configure.
const KiB = 1 << 10
const MiB = 1 << 20
const GiB = 1 << 30
const TiB = 1 << 40

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
			childSize, err := scanDir(
				childPath,
				folderMap,
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
