package main

import (
	"container/heap"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func sendErr(errCh chan<- error, err error) {
	if err == nil {
		return
	}
	errCh <- err
}

func scanDirConcurrent(
	fileSystem fs.FS,
	parentPath string,
	folderMap map[string]*DirNode,
	limit int,
	stats *ScanStats,
	wg *sync.WaitGroup,
	errCh chan<- error,
	mu *sync.Mutex,
) {
	defer wg.Done()
	h := &TupleHeap{}
	heap.Init(h)

	entries, err := fs.ReadDir(fileSystem, parentPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			mu.Lock()
			stats.PermissionSkip = append(stats.PermissionSkip, parentPath)
			mu.Unlock()
			return
		} else if errors.Is(err, fs.ErrNotExist) {
			mu.Lock()
			stats.NoDirSkip = append(stats.NoDirSkip, parentPath)
			mu.Unlock()
			return
		}
		sendErr(errCh, err)
		return
	}

	var totalSize int64
	var childPaths []string
	var childWG sync.WaitGroup
	for _, entry := range entries {
		if entry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		childPath := filepath.Join(parentPath, entry.Name())
		if entry.IsDir() {
			childPaths = append(childPaths, childPath)
			wg.Add(1)
			childWG.Add(1)
			go func(p string) {
				defer childWG.Done()
				scanDirConcurrent(
					fileSystem,
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
				if errors.Is(err, fs.ErrPermission) {
					mu.Lock()
					stats.PermissionSkip = append(stats.PermissionSkip, childPath)
					mu.Unlock()
					sendErr(errCh, err)
					continue
				}
				if errors.Is(err, fs.ErrNotExist) {
					mu.Lock()
					stats.NoDirSkip = append(stats.NoDirSkip, childPath)
					mu.Unlock()
					sendErr(errCh, err)
					continue
				}
				mu.Lock()
				stats.FileInfoSkip = append(stats.FileInfoSkip, childPath)
				mu.Unlock()
				continue
			}
			stats.TotalFileCount.Add(1)
			totalSize += info.Size()
		}
	}
	childWG.Wait()

	mu.Lock()
	for _, childPath := range childPaths {
		if child, ok := folderMap[childPath]; ok {
			totalSize += child.Size
			heap.Push(h, ChildTuple{Path: childPath, Size: child.Size})
		}
		if h.Len() > limit {
			heap.Pop(h)
		}
	}
	folderMap[parentPath] = &DirNode{
		FolderName:   filepath.Base(parentPath),
		Size:         totalSize,
		TopKChildren: showChildren(h),
	}
	mu.Unlock()

	fmt.Printf("Scanned %s: %s\n", filepath.Base(parentPath), sizeify(ByteSize(totalSize)))
}

func start_scanning_concurrent(fileSystem fs.FS, limit int) (map[string]*DirNode, *ScanStats, []error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Fresh state per run for fair timing
	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}
	start := time.Now()
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
		fileSystem,
		".",
		folderMap,
		limit,
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
		return nil, stats, errs
	}
	fmt.Printf("Scanning completed! Took %v to scan %d files. \n", time.Since(start), stats.TotalFileCount.Load())
	return folderMap, stats, nil
}
