package scanner

import (
	"os"
	"testing"

	"github.com/corona10/goimagehash"
)

func TestBKTree(t *testing.T) {
	// Create some dummy hashes
	// hash1 and hash2 will have distance 2
	// hash1 and hash3 will have distance 10
	h1 := goimagehash.NewImageHash(0, goimagehash.PHash)
	h2 := goimagehash.NewImageHash(3, goimagehash.PHash)    // bits 0 and 1 set
	h3 := goimagehash.NewImageHash(1023, goimagehash.PHash) // bits 0-9 set

	p1 := hashedPhoto{path: "p1", hash: h1, info: &dummyFileInfo{name: "p1"}}
	p2 := hashedPhoto{path: "p2", hash: h2, info: &dummyFileInfo{name: "p2"}}
	p3 := hashedPhoto{path: "p3", hash: h3, info: &dummyFileInfo{name: "p3"}}

	tree := NewBKTree()
	tree.Add(p1)
	tree.Add(p2)
	tree.Add(p3)

	// Search for p1 with distance 2
	results := tree.Search(h1, 2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	foundP1 := false
	foundP2 := false
	for _, p := range results {
		if p.path == "p1" {
			foundP1 = true
		}
		if p.path == "p2" {
			foundP2 = true
		}
	}

	if !foundP1 || !foundP2 {
		t.Errorf("Did not find expected photos: p1=%v, p2=%v", foundP1, foundP2)
	}

	// Search for p1 with distance 10
	results = tree.Search(h1, 10)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Search for something that matches nothing
	h4 := goimagehash.NewImageHash(0xFFFFFFFF00000000, goimagehash.PHash)
	results = tree.Search(h4, 0)
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

type dummyFileInfo struct {
	os.FileInfo
	name string
}

func (d *dummyFileInfo) Name() string { return d.name }
func (d *dummyFileInfo) Size() int64  { return 0 }
