package prefix_tree

// Node represents a node in the prefix tree.
type Node[T any] struct {
	right *Node[T]
	left  *Node[T]

	root     bool
	terminal bool
	value    *T // Can be nil
}

func NewNode[T any]() *Node[T] {
	return &Node[T]{terminal: false}
}

func RootNode[T any]() *Node[T] {
	node := NewNode[T]()
	node.root = true
	return node
}

func (n *Node[T]) IsRoot() bool {
	return n.root
}

func (n *Node[T]) IsLeaf() bool {
	return nil == n.right && nil == n.left
}

func (n *Node[T]) IsTerminal() bool {
	return n.terminal
}

func (n *Node[T]) MarkTerminal() {
	n.terminal = true
}

func (n *Node[T]) UnmarkTerminal() {
	n.terminal = false
}

func (n *Node[T]) SaveAndMarkTerminal(value *T) {
	n.value = value
	n.MarkTerminal()
}

// treeNodeStack is a simple stack implementation for treeNode pointers.
// Used to assist in tree traversals.
type NodeStack[T any] struct {
	nodes []*Node[T]
}

func NewNodeStack[T any]() *NodeStack[T] {
	return &NodeStack[T]{
		nodes: make([]*Node[T], 0),
	}
}

func (ns *NodeStack[T]) Push(node *Node[T]) {
	ns.nodes = append(ns.nodes, node)
}

func (ns *NodeStack[T]) Pop() *Node[T] {
	if len(ns.nodes) == 0 {
		return nil
	}

	node := ns.nodes[len(ns.nodes)-1]
	ns.nodes = ns.nodes[:len(ns.nodes)-1]
	return node
}

func (ns *NodeStack[T]) IsEmpty() bool {
	return len(ns.nodes) == 0
}

func (ns *NodeStack[T]) Size() int {
	return len(ns.nodes)
}
