package prefix_tree

// Node represents a node in the prefix tree.
type Node[T any] struct {
	right *Node[T]
	left  *Node[T]

	terminal bool
	value    T // Can be nil
}

// Root node. Same as Node.
type RootNode[T any] struct {
	*Node[T]
}

func NewNode[T any]() *Node[T] {
	return &Node[T]{terminal: false}
}

func NewRootNode[T any]() *RootNode[T] {
	return &RootNode[T]{
		Node: NewNode[T](),
	}
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

func (n *Node[T]) SaveAndMarkTerminal(value T) {
	n.value = value
	n.MarkTerminal()
}

// treeNodeStack is a simple stack implementation for treeNode pointers.
// Used to assist in tree traversals.
type NodeStack[T any] struct {
	nodes []*Node[T]
}

// Creates a new NodeStack
func NewNodeStack[T any]() *NodeStack[T] {
	return &NodeStack[T]{
		nodes: make([]*Node[T], 0),
	}
}

// Pushes a node onto the stack
func (ns *NodeStack[T]) Push(node *Node[T]) {
	ns.nodes = append(ns.nodes, node)
}

// Pops a node from the stack
func (ns *NodeStack[T]) Pop() *Node[T] {
	if len(ns.nodes) == 0 {
		return nil
	}

	node := ns.nodes[len(ns.nodes)-1]
	ns.nodes = ns.nodes[:len(ns.nodes)-1]
	return node
}

// Peek returns the top node without removing it from the stack
func (ns *NodeStack[T]) Peek() *Node[T] {
	if len(ns.nodes) == 0 {
		return nil
	}

	return ns.nodes[len(ns.nodes)-1]
}

// Checks if the stack is empty
func (ns *NodeStack[T]) IsEmpty() bool {
	return len(ns.nodes) == 0
}

// Returns the size of the stack
func (ns *NodeStack[T]) Size() int {
	return len(ns.nodes)
}
