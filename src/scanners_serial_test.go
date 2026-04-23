package main

import (
	"container/heap"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
)

type denyReadDirFS struct {
	fsys fs.FS
	deny map[string]error
}

func (d denyReadDirFS) Open(name string) (fs.File, error) {
	return d.fsys.Open(name)
}

func (d denyReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	clean := filepath.Clean(name)
	if err, ok := d.deny[clean]; ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}
	if rdfs, ok := d.fsys.(fs.ReadDirFS); ok {
		return rdfs.ReadDir(name)
	}
	return fs.ReadDir(d.fsys, name)
}

func TestShowChildren(t *testing.T) {
	h := &TupleHeap{}
	heap.Init(h)

	heap.Push(h, ChildTuple{Path: "a", Size: 1})
	heap.Push(h, ChildTuple{Path: "b", Size: 3})
	heap.Push(h, ChildTuple{Path: "c", Size: 2})

	got := showChildren(h)
	want := []string{"b", "c", "a"}

	equals(t, want, got)
}

func TestScanDir_BasicCountsAndTopK(t *testing.T) {
	fsys := fstest.MapFS{
		"file1":    {Data: make([]byte, 100)},
		"a/f2":     {Data: make([]byte, 200)},
		"b/f3":     {Data: make([]byte, 50)},
		"b/f4":     {Data: make([]byte, 1)},
		"a/nested": {Mode: fs.ModeDir},
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	total, err := scanDir(fsys, ".", folderMap, 1, stats)

	ok(t, err)
	equals(t, int64(351), total)

	root := folderMap["."]
	assert(t, root != nil, "Missing node for '.'")
	equals(t, int64(351), root.Size)
	equals(t, []string{"a"}, root.TopKChildren)
	assert(t, folderMap["a"] != nil, "Node 'a' is missing.")
	equals(t, int64(200), folderMap["a"].Size)
	assert(t, folderMap["b"] != nil, "Node 'b' is missing.")
	equals(t, int64(51), folderMap["b"].Size)
}

func TestScanDir_SymlinkSkipped(t *testing.T) {
	fsys := fstest.MapFS{
		"real": {Data: make([]byte, 10)},
		"sym":  {Mode: fs.ModeSymlink, Data: make([]byte, 999)},
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	total, err := scanDir(fsys, ".", folderMap, 10, stats)
	ok(t, err)
	equals(t, int64(10), total)
}

func TestScanDir_PermissionDenied(t *testing.T) {
	base := fstest.MapFS{
		"public/ok":        {Data: make([]byte, 5)},
		"private/secret":   {Data: make([]byte, 1000)},
		"private2/secret2": {Data: make([]byte, 1000)},
	}
	fsys := denyReadDirFS{
		fsys: base,
		deny: map[string]error{
			"private":  fs.ErrPermission,
			"private2": fs.ErrPermission,
		},
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	total, err := scanDir(fsys, ".", folderMap, 10, stats)

	assert(t, err == nil, "scanDir error: ", err)
	assert(t, total == 5, "total=%d, want 5 (private should be skipped)", total)
	assert(t, filepath.Clean(stats.PermissionSkip[0]) == "private", "stats.Skipped=%v, want [private]", stats.PermissionSkip)
	assert(t, len(stats.PermissionSkip) == 2, "stats Skipped.", "stats.Skipped=%v, want %v", stats.PermissionSkip, 2)
}

func TestScanDir_NonExistentChildIsSkipped(t *testing.T) {
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

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	total, err := scanDir(fsys, ".", folderMap, 10, stats)
	ok(t, err)
	equals(t, int64(7), total)
	equals(t, 1, len(stats.NoDirSkip))
	equals(t, "gone", filepath.Clean(stats.NoDirSkip[0]))
}

func TestStart_Scanning_WithMapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"p/q":   {Data: make([]byte, 3)},
		"r/s/t": {Data: make([]byte, 4)},
	}

	// IMPORTANT: for fs.FS (including fstest.MapFS and os.DirFS),
	// the "root" path inside the FS is typically ".".
	folderMap, _, err := start_scanning(fsys, 10)
	ok(t, err)
	assert(t, folderMap["."] != nil, "expected folderMap to contain '.' root node")
	assert(t, folderMap["."].Size == int64(7), "root size=%d, want 7", folderMap["."].Size)
}

func TestScanDir_ErrorIsWrapped(t *testing.T) {
	// Ensure error wrapping still preserves Is/As behavior.
	fsys := denyReadDirFS{
		fsys: fstest.MapFS{},
		deny: map[string]error{
			".": fs.ErrPermission,
		},
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	_, err := scanDir(fsys, ".", folderMap, 10, stats)
	assert(t, err != nil, "expected a permission error to occur")
	assert(t, errors.Is(err, fs.ErrPermission), "expected errors.Is(err, fs.ErrPermission)=true, got %v", err)
}
