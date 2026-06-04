package scanner

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
	if t.root == nil {
		t.root = &bkTreeNode{
			photo:    photo,
			children: make(map[int]*bkTreeNode),
		}
		return
	}

	curr := t.root
	for {
		distance, _ := photo.hash.Distance(curr.photo.hash)

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
// It appends results to the provided slice to avoid allocations.
func (t *BKTree) Search(hash PHash, maxDistance int, results *[]hashedPhoto) {
	if t.root == nil {
		return
	}

	// Use a small local array for the queue to avoid heap allocation for small trees
	// For larger trees, this may still grow, but it's a significant improvement.
	queue := make([]*bkTreeNode, 0, 16)
	queue = append(queue, t.root)

	for len(queue) > 0 {
		curr := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		distance, _ := hash.Distance(curr.photo.hash)

		if distance <= maxDistance {
			*results = append(*results, curr.photo)
		}

		minD := distance - maxDistance
		maxD := distance + maxDistance

		for d, child := range curr.children {
			if d >= minD && d <= maxD {
				queue = append(queue, child)
			}
		}
	}
}
