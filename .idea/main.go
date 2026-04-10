package main

import (
	"container/heap"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

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

func sendErr(errCh chan<- error, err error) {
	if err == nil {
		return
	}
	errCh <- err
}

func scanDirConcurrent(
	parentPath string,
	folderMap map[string]*DirNode,
	limit int,
	stats *ScanStats,
	wg *sync.WaitGroup,
	errCh chan<- error,
	mu *sync.Mutex,
) {
	defer wg.Done()

	entries, err := os.ReadDir(parentPath)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			mu.Lock()
			stats.Skipped = append(stats.Skipped, parentPath)
			mu.Unlock()
			return
		}
		sendErr(errCh, err)
		return
	}

	var size int64
	var childPaths []string
	var childWG sync.WaitGroup
	for _, entry := range entries {
		childPath := filepath.Join(parentPath, entry.Name())
		if entry.IsDir() {
			childPaths = append(childPaths, childPath)
			wg.Add(1)
			childWG.Add(1)
			go func(p string) {
				defer childWG.Done()
				scanDirConcurrent(
					p,
					folderMap,
					limit,
					stats,
					wg,
					errCh,
					mu,
				)
			}(childPath)
		} else {
			info, err := entry.Info()
			if err != nil {
				sendErr(errCh, err)
				continue
			}
			size += info.Size()
		}
	}
	childWG.Wait()

	mu.Lock()
	for _, p := range childPaths {
		if n, ok := folderMap[p]; ok {
			size += n.Size
		}
	}
	folderMap[parentPath] = &DirNode{
		FolderName: filepath.Base(parentPath),
		Size:       size,
	}
	mu.Unlock()

	fmt.Printf("Scanned %s: %d\n", filepath.Base(parentPath), size)
}

func start_scanning_concurrent() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	const runs = 10
	for range runs {

		// Fresh state per run for fair timing
		folderMap := make(map[string]*DirNode)
		stats := &ScanStats{}
		errCh := make(chan error)

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errsMu sync.Mutex

		errs := make([]error, 0, 16)
		collectDone := make(chan struct{})
		go func() {
			defer close(collectDone)
			for err := range errCh {
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
			}
		}()
		wg.Add(1)
		go scanDirConcurrent(
			"/Users/Tohya",
			folderMap,
			10,
			stats,
			&wg,
			errCh,
			&mu,
		)

		go func() {
			wg.Wait()
			close(errCh)
		}()

		<-collectDone

		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Println("Error occurred:", err)
			}
			os.Exit(1)
		}
	}
}

func main() {

}
