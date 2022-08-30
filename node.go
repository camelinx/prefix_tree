package prefix_tree

type treeNode struct {
    right       *treeNode
    left        *treeNode
    parent      *treeNode

    terminal     bool
    value        interface{ }
}

func newNode( )( *treeNode ) {
    return &treeNode{ terminal: false }
}

func ( node *treeNode )isRoot( )( bool ) {
    return nil != node && nil == node.parent
}

func ( node *treeNode )isLeaf( )( bool ) {
    return nil != node && nil == node.right && nil == node.left
}

func ( node *treeNode )isTerminal( )( bool ) {
    return nil != node && node.terminal
}

func ( node *treeNode )markTerminal( )( ) {
    if nil != node {
        node.terminal = true
    }
}
