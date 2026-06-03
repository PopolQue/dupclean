package diskanalyzer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderByType(t *testing.T) {
	result := &AnalysisResult{
		TotalSize: 1000,
		AllFiles: []FileEntry{
			{Name: "f1.txt", Size: 600, Ext: ".txt"},
			{Name: "f2.log", Size: 400, Ext: ".log"},
		},
	}

	buf := new(bytes.Buffer)
	renderByType(buf, result)

	output := buf.String()
	if !strings.Contains(output, ".txt") || !strings.Contains(output, "60.0%") {
		t.Errorf("renderByType output incorrect: %s", output)
	}
}

func TestRenderCLI(t *testing.T) {
	root := &DirNode{Name: "root", Path: "/", TotalSize: 1024}
	result := &AnalysisResult{
		Root:      root,
		TotalSize: 1024,
	}
	opts := CLIOptions{Depth: 1}

	buf := new(bytes.Buffer)
	RenderCLI(buf, result, opts)

	if !strings.Contains(buf.String(), "/  —  1.0 KB") {
		t.Errorf("RenderCLI output missing header: %s", buf.String())
	}
}

func TestRenderTopFiles(t *testing.T) {
	result := &AnalysisResult{
		AllFiles: []FileEntry{
			{Name: "f1.txt", Size: 600, Path: "/path/f1.txt"},
		},
	}

	buf := new(bytes.Buffer)
	renderTopFiles(buf, result, 1)

	if !strings.Contains(buf.String(), "f1.txt") {
		t.Errorf("renderTopFiles output incorrect: %s", buf.String())
	}
}

func TestRenderOldFiles(t *testing.T) {
	result := &AnalysisResult{
		AllFiles: []FileEntry{
			{Name: "f1.txt", Size: 600, Path: "/path/f1.txt", ModTime: time.Now().AddDate(0, 0, -10)},
		},
	}

	buf := new(bytes.Buffer)
	renderOldFiles(buf, result, 5, 100) // Older than 5 days, min 100 bytes

	if !strings.Contains(buf.String(), "f1.txt") {
		t.Errorf("renderOldFiles output incorrect: %s", buf.String())
	}
}

func TestRenderTree(t *testing.T) {
	root := &DirNode{Name: "root", Path: "/", TotalSize: 1000}
	root.Files = []FileEntry{
		{Name: "f1.txt", Size: 1000},
	}

	buf := new(bytes.Buffer)
	renderTree(buf, root, 0, 1, 1000)

	if !strings.Contains(buf.String(), "f1.txt") {
		t.Errorf("renderTree output incorrect: %s", buf.String())
	}
}
