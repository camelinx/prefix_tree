package prefix_tree

import (
	"testing"
)

func TestTreeNode(t *testing.T) {
	node := NewNode[string]()
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

func TestTreeNodeStack(t *testing.T) {
	stack := NewNodeStack[string]()
	if stack == nil {
		t.Fatalf("NewNodeStack: returned nil stack")
	}

	if !stack.IsEmpty() {
		t.Fatalf("IsEmpty: stack incorrectly identified as non-empty")
	}

	if stack.Peek() != nil {
		t.Fatalf("Peek: expected nil on empty stack, got %v", stack.Peek())
	}

	if stack.Size() != 0 {
		t.Fatalf("Len: expected size 0 on empty stack, got %d", stack.Size())
	}

	node1 := NewNode[string]()
	node2 := NewNode[string]()

	stack.Push(node1)
	if stack.IsEmpty() {
		t.Fatalf("IsEmpty: stack incorrectly identified as empty")
	}
	if stack.Peek() != node1 {
		t.Fatalf("Peek: top of stack does not match expected node1")
	}

	stack.Push(node2)
	if stack.IsEmpty() {
		t.Fatalf("IsEmpty: stack incorrectly identified as empty")
	}
	if stack.Peek() != node2 {
		t.Fatalf("Peek: top of stack does not match expected node2")
	}
	if stack.Size() != 2 {
		t.Fatalf("Len: stack length incorrect, expected 2 got %d", stack.Size())
	}

	poppped := stack.Pop()
	if poppped != node2 {
		t.Fatalf("Pop: popped node does not match expected node2")
	}
	if stack.Size() != 1 {
		t.Fatalf("Len: stack length incorrect after pop, expected 1 got %d", stack.Size())
	}

	poppped = stack.Pop()
	if poppped != node1 {
		t.Fatalf("Pop: popped node does not match expected node1")
	}
	if !stack.IsEmpty() {
		t.Fatalf("IsEmpty: stack incorrectly identified as non-empty after pops")
	}

	poppped = stack.Pop()
	if poppped != nil {
		t.Fatalf("Pop: expected nil when popping from empty stack, got %v", poppped)
	}
}
