package prefix_tree

import (
	"testing"
)

func TestTreeNode(t *testing.T) {
	node := RootNode[string]()
	if !node.IsRoot() {
		t.Fatalf("isRoot: failed to recognize root node")
	}

	node = NewNode[string]()
	node.root = false
	if node.IsRoot() {
		t.Fatalf("isRoot: incorrectly identified node as root")
	}

	if !node.IsLeaf() {
		t.Fatalf("isLeaf: failed to recognize leaf node")
	}

	node.right = NewNode[string]()
	if node.IsLeaf() {
		t.Fatalf("isLeaf: incorrectly identified node as leaf")
	}

	node.left = NewNode[string]()
	if node.IsLeaf() {
		t.Fatalf("isLeaf: incorrectly identified node as leaf")
	}

	if node.IsTerminal() {
		t.Fatalf("isTerminal: incorrectly identified node a terminal")
	}

	node.MarkTerminal()
	if !node.IsTerminal() {
		t.Fatalf("isTerminal: failed to recognize terminal node")
	}

	node.UnmarkTerminal()
	if node.IsTerminal() {
		t.Fatalf("isTerminal: incorrectly identified node a terminal")
	}
}
