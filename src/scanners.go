package main

import (
	"container/heap"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
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
	PermissionSkip []string
	NoDirSkip      []string
	FileInfoSkip   []string
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

// List topK largest subdirectories in reverse-size order.
func showChildren(h *TupleHeap) []string {
	n := h.Len()
	out := make([]string, n)
	for i := n - 1; i >= 0; i-- {
		out[i] = heap.Pop(h).(ChildTuple).Path
	}

	return out
}

// Recursive bottom-up approach to compute each directory size
// and its topK largest subdirectories
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
			stats.PermissionSkip = append(stats.PermissionSkip, parentPath)
			return 0, fmt.Errorf("skipped permission denied for the path %q:  %w", parentPath, err)
		} else if errors.Is(err, fs.ErrNotExist) {
			stats.NoDirSkip = append(stats.NoDirSkip, parentPath)
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
			// Use a classic maxHeap trick for minHeap
			heap.Push(h, ChildTuple{Path: childPath, Size: childSize})
			if h.Len() > limit {
				heap.Pop(h)
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				if errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrNotExist) {
					continue
				}
				stats.FileInfoSkip = append(stats.FileInfoSkip, childPath)
				return 0, fmt.Errorf("unexpected error when opening a file %q: %w", info, err)
			}
			totalSize += info.Size()
		}
	}
	folderMap[parentPath] = &DirNode{
		FolderName:   filepath.Base(parentPath),
		Size:         totalSize,
		TopKChildren: showChildren(h),
	}

	fmt.Printf("Scanned %s: %s\n", filepath.Base(parentPath), sizeify(ByteSize(totalSize)))
	return totalSize, nil
}

// Start filescanning. limit is the number of largest subdirs to show.
func startScanning(fileSystem fs.FS, limit int) (map[string]*DirNode, error) {
	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}
	start := time.Now()
	if _, err := scanDir(fileSystem, ".", folderMap, limit, stats); err != nil {
		// Return the wrapped error
		time.Since(start)
		return nil, err
	}
	if len(stats.PermissionSkip)+len(stats.NoDirSkip) > 0 {
		fmt.Printf("\nSkipped %d directories/files due to permission denied and %d directories/files due to non-existent files. Consider running 'sudo hddviz' to bypass the permission issue.\n", len(stats.PermissionSkip), len(stats.NoDirSkip))
	}
	fmt.Printf("Scanning completed! Took %v to run. \n", time.Since(start))
	return folderMap, nil
}
