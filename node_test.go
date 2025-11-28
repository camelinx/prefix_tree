package prefix_tree

import (
	"testing"
)

func TestTreeNode(t *testing.T) {
	node := rootNode()
	if !node.isRoot() {
		t.Fatalf("isRoot: failed to recognize root node")
	}

	node = newNode()
	node.root = false
	if node.isRoot() {
		t.Fatalf("isRoot: incorrectly identified node as root")
	}

	if !node.isLeaf() {
		t.Fatalf("isLeaf: failed to recognize leaf node")
	}

	node.right = newNode()
	if node.isLeaf() {
		t.Fatalf("isLeaf: incorrectly identified node as leaf")
	}

	node.left = newNode()
	if node.isLeaf() {
		t.Fatalf("isLeaf: incorrectly identified node as leaf")
	}

	if node.isTerminal() {
		t.Fatalf("isTerminal: incorrectly identified node a terminal")
	}

	node.markTerminal()
	if !node.isTerminal() {
		t.Fatalf("isTerminal: failed to recognize terminal node")
	}

	node.unmarkTerminal()
	if node.isTerminal() {
		t.Fatalf("isTerminal: incorrectly identified node a terminal")
	}

	node = nil
	node.isRoot()
	node.isLeaf()
	node.isTerminal()
	node.markTerminal()
	node.unmarkTerminal()
}
