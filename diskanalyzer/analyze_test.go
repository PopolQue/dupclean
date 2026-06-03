package diskanalyzer

import (
	"testing"
	"time"
)

func TestAnalysisLogic(t *testing.T) {
	root := &DirNode{Name: "root", Path: "/", TotalSize: 1000}
	child := &DirNode{Name: "child", Path: "/child", TotalSize: 400, Parent: root}
	root.Children = []*DirNode{child}

	files := []FileEntry{
		{Name: "f1", Size: 600, Ext: ".txt", ModTime: time.Now().AddDate(0, 0, -10)},
		{Name: "f2", Size: 400, Ext: ".log", ModTime: time.Now().AddDate(0, 0, -5)},
	}

	result := &AnalysisResult{
		Root:      root,
		AllFiles:  files,
		TotalSize: 1000,
	}

	// Test TopFiles
	if len(TopFiles(result, 0)) != 0 {
		t.Error("TopFiles(0) should return 0 items")
	}
	if len(TopFiles(result, 2)) != 2 {
		t.Error("TopFiles(2) should return all 2 items")
	}

	// Test TypeBreakdown
	breakdown := TypeBreakdown(result)
	found := false
	for _, b := range breakdown {
		if b.Ext == "(none)" {
			t.Error("Should not have empty extension")
		}
		if b.Ext == ".txt" && b.PctOfDisk != 60.0 {
			t.Errorf("Wrong pct for .txt: %f", b.PctOfDisk)
		}
		found = true
	}
	if !found {
		t.Error("Breakdown missing")
	}

	// Test LargestDirs
	if len(LargestDirs(result, 0)) != 2 {
		t.Error("LargestDirs(0) should return all dirs")
	}
	if len(LargestDirs(result, -1)) != 2 {
		t.Error("LargestDirs(-1) should return all dirs")
	}

	// Test OldFiles
	old := OldFiles(result, 7, 100)
	if len(old) != 1 || old[0].Name != "f1" {
		t.Errorf("OldFiles failed, got: %v", old)
	}
}
