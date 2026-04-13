package main

import (
	"container/heap"
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
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

func TestFilterChildren(t *testing.T) {
	h := &TupleHeap{}
	heap.Init(h)

	heap.Push(h, ChildTuple{Path: "a", Size: 1})
	heap.Push(h, ChildTuple{Path: "b", Size: 3})
	heap.Push(h, ChildTuple{Path: "c", Size: 2})

	got := filterChildren(h)
	want := []string{"b", "c", "a"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterChildren()=%v, want %v", got, want)
	}
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
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if total != 351 {
		t.Fatalf("total=%d, want 351", total)
	}

	root := folderMap["."]
	if root == nil {
		t.Fatalf("missing node for '.'")
	}
	if root.Size != 351 {
		t.Fatalf("root.Size=%d, want 351", root.Size)
	}
	if len(root.TopKChildren) != 1 || root.TopKChildren[0] != "a" {
		t.Fatalf("root.TopKChildren=%v, want [a]", root.TopKChildren)
	}

	if folderMap["a"] == nil || folderMap["a"].Size != 200 {
		t.Fatalf("node 'a' missing or wrong size: %#v", folderMap["a"])
	}
	if folderMap["b"] == nil || folderMap["b"].Size != 51 {
		t.Fatalf("node 'b' missing or wrong size: %#v", folderMap["b"])
	}
}

func TestScanDir_SymlinkSkipped(t *testing.T) {
	fsys := fstest.MapFS{
		"real": {Data: make([]byte, 10)},
		"sym":  {Mode: fs.ModeSymlink, Data: make([]byte, 999)},
	}

	folderMap := make(map[string]*DirNode)
	stats := &ScanStats{}

	total, err := scanDir(fsys, ".", folderMap, 10, stats)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if total != 10 {
		t.Fatalf("total=%d, want 10 (symlink should be skipped)", total)
	}
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
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if total != 5 {
		t.Fatalf("total=%d, want 5 (private should be skipped)", total)
	}
	if len(stats.Skipped) != 2 || filepath.Clean(stats.Skipped[0]) != "private" {
		t.Fatalf("stats.Skipped=%v, want [private]", stats.Skipped)
	}
}

func TestScanDir_NonExistentChildIsSkipped(t *testing.T) {
	base := fstest.MapFS{
		"keep/file": {Data: make([]byte, 7)},
		// NOTE: no "gone/..." keys => directory effectively doesn't exist.
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
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if total != 7 {
		t.Fatalf("total=%d, want 7 (gone should be skipped)", total)
	}
	if len(stats.Skipped) != 1 || filepath.Clean(stats.Skipped[0]) != "gone" {
		t.Fatalf("stats.Skipped=%v, want [gone]", stats.Skipped)
	}
}

func TestStartScanning_WithMapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"p/q":   {Data: make([]byte, 3)},
		"r/s/t": {Data: make([]byte, 4)},
	}

	// IMPORTANT: for fs.FS (including fstest.MapFS and os.DirFS),
	// the "root" path inside the FS is typically ".".
	folderMap, err := startScanning(fsys, ".", 10)
	if err != nil {
		t.Fatalf("startScanning error: %v", err)
	}

	if folderMap["."] == nil {
		t.Fatalf("expected folderMap to contain '.' root node")
	}
	if folderMap["."].Size != 7 {
		t.Fatalf("root size=%d, want 7", folderMap["."].Size)
	}
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
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("expected errors.Is(err, fs.ErrPermission)=true, got %v", err)
	}
}
