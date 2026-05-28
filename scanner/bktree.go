package scanner

import (
	"github.com/corona10/goimagehash"
)

// bkTreeNode represents a node in the BK-Tree
type bkTreeNode struct {
	photo    hashedPhoto
	children map[int]*bkTreeNode
}

// BKTree implements a Burkhard-Keller Tree for efficient similarity searching
// based on Hamming distance.
type BKTree struct {
	root *bkTreeNode
}

// NewBKTree creates a new empty BK-Tree
func NewBKTree() *BKTree {
	return &BKTree{}
}

// Add inserts a hashed photo into the BK-Tree
func (t *BKTree) Add(photo hashedPhoto) {
	if photo.hash == nil {
		return
	}

	if t.root == nil {
		t.root = &bkTreeNode{
			photo:    photo,
			children: make(map[int]*bkTreeNode),
		}
		return
	}

	curr := t.root
	for {
		distance, err := photo.hash.Distance(curr.photo.hash)
		if err != nil {
			// In practice, goimagehash.Distance only fails if hash types differ
			return
		}

		// If distance is 0, we could either store multiple photos at the node
		// or just treat them as identical. For our purposes, we can have nodes
		// with distance 0 if we want, but usually Hamming distance > 0 for different nodes.
		// If distance is 0, they are essentially the same hash.
		if next, ok := curr.children[distance]; ok {
			curr = next
		} else {
			curr.children[distance] = &bkTreeNode{
				photo:    photo,
				children: make(map[int]*bkTreeNode),
			}
			break
		}
	}
}

// Search finds all photos in the tree within the given maximum Hamming distance
func (t *BKTree) Search(hash *goimagehash.ImageHash, maxDistance int) []hashedPhoto {
	if t.root == nil || hash == nil {
		return nil
	}

	var results []hashedPhoto
	queue := []*bkTreeNode{t.root}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		distance, err := hash.Distance(curr.photo.hash)
		if err != nil {
			continue
		}

		if distance <= maxDistance {
			results = append(results, curr.photo)
		}

		// Only search children that could possibly contain matches
		// according to the triangle inequality:
		// distance(curr, target) = d
		// For any child node at distance(curr, child) = d_child
		// matches must satisfy |d - d_child| <= maxDistance
		// so d - maxDistance <= d_child <= d + maxDistance
		minD := distance - maxDistance
		maxD := distance + maxDistance

		for d, child := range curr.children {
			if d >= minD && d <= maxD {
				queue = append(queue, child)
			}
		}
	}

	return results
}
