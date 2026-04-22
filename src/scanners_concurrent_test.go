package main

import (
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestScanDirConcurrent(t *testing.T) {
	fsys := fstest.MapFS{
		"file1":   {Data: make([]byte, 50)},
		"a/file2": {Data: make([]byte, 20)},
		"b/file3": {Data: make([]byte, 10)},
	}

	folderMap, stats, errs := start_scanning_concurrent(fsys, 1)
	assert(t, len(errs) == 0, "Concurrent scanning failed")
	assert(t, stats.TotalFileCount.Load() == uint64(3), "Undercounting files", stats.TotalFileCount.Load())

	root := folderMap["."]
	assert(t, root != nil, "Missing node for '.'")
	equals(t, int64(80), root.Size)
	equals(t, []string{"a"}, root.TopKChildren)
	assert(t, folderMap["a"] != nil, "Node 'a' is missing.")
	equals(t, int64(20), folderMap["a"].Size)
	assert(t, folderMap["b"] != nil, "Node 'b' is missing.")
	equals(t, int64(10), folderMap["b"].Size)
}

func TestScanDirConcurrent_SymlinkSkipped(t *testing.T) {
	fsys := fstest.MapFS{
		"real": {Data: make([]byte, 10)},
		"sym":  {Mode: fs.ModeSymlink, Data: make([]byte, 999)},
	}
	folderMap, _, errs := start_scanning_concurrent(fsys, 1)
	assert(t, len(errs) == 0, "Concurrent scanning failed")
	equals(t, int64(10), folderMap["."].Size)
}

func TestScanDirConcurrent_PermissionDenieed(t *testing.T) {
	base_fsys := fstest.MapFS{
		"public/nice":      {Data: make([]byte, 5)},
		"private/secret":   {Data: make([]byte, 1000)},
		"private2/secret2": {Data: make([]byte, 1000)},
	}
	fsys := denyReadDirFS{
		fsys: base_fsys,
		deny: map[string]error{
			"private":  fs.ErrPermission,
			"private2": fs.ErrPermission,
		},
	}

	folderMap, _, errs := start_scanning_concurrent(fsys, 1)
	assert(t, len(errs) == 0, "Concurrent scanning failed")
	equals(t, int64(5), folderMap["."].Size)
}

func TestScanDirConcurrent_NonExistentChildIsSkipped(t *testing.T) {
	base := fstest.MapFS{
		"keep/file": {Data: make([]byte, 7)},
	}
	fsys := denyReadDirFS{
		fsys: base,
		deny: map[string]error{
			"gone": fs.ErrNotExist,
		},
	}

	// Make root listing include "gone" by creating a descendant key, then denying readdir.
	// Without this, MapFS won't necessarily list "gone" as a directory.
	base["gone/x"] = &fstest.MapFile{Data: make([]byte, 1)}

	folderMap, stats, errors := start_scanning_concurrent(fsys, 1)
	assert(t, len(errors) == 0, "No errors should exist", errors)
	equals(t, int64(7), folderMap["."].Size)
	equals(t, 1, len(stats.NoDirSkip))
	equals(t, "gone", filepath.Clean(stats.NoDirSkip[0]))
}
