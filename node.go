package prefix_tree

// treeNode represents a node in the prefix tree.
type treeNode struct {
	right *treeNode
	left  *treeNode

	root     bool
	terminal bool
	value    interface{}
}

func newNode() *treeNode {
	return &treeNode{terminal: false}
}

func rootNode() *treeNode {
	node := newNode()
	node.root = true
	return node
}

func (node *treeNode) isRoot() bool {
	return nil != node && node.root
}

func (node *treeNode) isLeaf() bool {
	return nil != node && nil == node.right && nil == node.left
}

func (node *treeNode) isTerminal() bool {
	return nil != node && node.terminal
}

func (node *treeNode) markTerminal() {
	if nil != node {
		node.terminal = true
	}
}

func (node *treeNode) unmarkTerminal() {
	if nil != node {
		node.terminal = false
	}
}

func (node *treeNode) saveAndMarkTerminal(value interface{}) {
	node.value = value
	node.markTerminal()
}

// treeNodeStack is a simple stack implementation for treeNode pointers.
// Used to assist in tree traversals.
type treeNodeStack struct {
	nodes []*treeNode
}

func newTreeNodeStack() *treeNodeStack {
	return &treeNodeStack{
		nodes: make([]*treeNode, 0),
	}
}

func (s *treeNodeStack) Push(node *treeNode) {
	s.nodes = append(s.nodes, node)
}

func (s *treeNodeStack) Pop() *treeNode {
	if len(s.nodes) == 0 {
		return nil
	}

	node := s.nodes[len(s.nodes)-1]
	s.nodes = s.nodes[:len(s.nodes)-1]
	return node
}

func (s *treeNodeStack) IsEmpty() bool {
	return len(s.nodes) == 0
}

func (s *treeNodeStack) Size() int {
	return len(s.nodes)
}
