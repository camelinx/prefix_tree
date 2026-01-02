package prefix_tree

// Implements a generic prefix tree (trie) data structure. Implements insert, delete and exact/partial search operations.
// The tree can store any type of value associated with the keys. The keys are represented as byte slices along with
// corresponding masks. Masks are useful when storing IP addresses in CIDR notation. The tree supports lock handlers
// for concurrent access.
//
// TODO:
//   1. Tree compression to optimize memory usage and performance.
//   2. Iterator to traverse the tree.

import (
	"context"
	"fmt"
)

type Tree[T any] struct {
	root *RootNode[T]

	numNodes uint64

	rlockFn   ReadLockFn
	runlockFn ReadUnlockFn
	wlockFn   WriteLockFn
	unlockFn  UnlockFn
}

// Walker function
type TreeWalkerFn[T any] func(context.Context, T) error

// Returns a new prefix tree
// Returns:
//
//	*Tree - pointer to the new prefix tree
func NewTree[T any]() *Tree[T] {
	return &Tree[T]{
		root:     NewRootNode[T](),
		numNodes: 0,
	}
}

// Returns a new prefix tree with lock handlers set
// Arguments:
//
//	rlockFn   - read lock function
//	runlockFn - read unlock function
//	wlockFn   - write lock function
//	unlockFn  - unlock function
//
// Returns:
//
//	*Tree - pointer to the new prefix tree
func NewTreeWithLockHandlers[T any](rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) *Tree[T] {
	t := NewTree[T]()
	t.rlockFn = rlockFn
	t.runlockFn = runlockFn
	t.wlockFn = wlockFn
	t.unlockFn = unlockFn
	return t
}

func (t *Tree[T]) IsRoot(node *Node[T]) bool {
	return t.root.Node == node
}

func (t *Tree[T]) IsEmpty() bool {
	return t.numNodes == 0
}

func (t *Tree[T]) rlock(ctx context.Context) {
	if t.rlockFn != nil {
		t.rlockFn(ctx)
	}
}

func (t *Tree[T]) runlock(ctx context.Context) {
	if t.runlockFn != nil {
		t.runlockFn(ctx)
	}
}

func (t *Tree[T]) wlock(ctx context.Context) {
	if t.wlockFn != nil {
		t.wlockFn(ctx)
	}
}

func (t *Tree[T]) unlock(ctx context.Context) {
	if t.unlockFn != nil {
		t.unlockFn(ctx)
	}
}

func (t *Tree[T]) incrNumNodes() {
	t.numNodes++
}

func (t *Tree[T]) decrNumNodes() {
	if t.numNodes > 0 {
		t.numNodes--
	}
}

var (
	msbByteVal byte = byte(0x80) // 1000 0000
)

// Insert a key into the prefix tree. Will write lock the tree when inserting.
// Arguments:
//
//		ctx  - context for the lock functions.
//		key  - key to insert expressed as byte slice.
//		mask - mask for the key expressed as byte slice. A mask with non-contiguous
//			   1s is considered unexpected and will lead to undefined behavior.
//	           The very first bit of mask cannot be 0. The only way to store this node is to mark
//	           the root as terminal. This is not supported.
//		value - value associated with the key. This is optional and can be nil.
//
// Returns:
//
//	OpResult - result of the operation
//	error    - error if any
func (t *Tree[T]) Insert(ctx context.Context, key []byte, mask []byte, value T) (OpResult, error) {
	// key and mask lengths must be the same
	if len(key) != len(mask) {
		return Error, ErrInvalidKeyMask
	}

	keyLen := len(key)
	if keyLen <= 0 {
		return Error, fmt.Errorf("invalid key length %d", keyLen)
	}

	maskIdx := 0
	match := msbByteVal

	// The very first bit of mask cannot be 0
	// The only way to store this node is to mark
	// the root as terminal. This is not supported.
	if match != match&mask[maskIdx] {
		return Error, ErrInvalidKeyMask
	}

	t.wlock(ctx)
	defer func() {
		t.unlock(ctx)
	}()

	// Start from root
	node := t.root.Node
	next := t.root.Node

	// Traverse down the tree as far as possible.
	// Note: the first occurence of 0 in the mask will terminate the traversal.
	// It is assumed that all bits after the first 0 in the mask are also 0s.
	// There is no explicit check for this condition here. A mask with non-contiguous
	// 1s is considered unexpected and will lead to undefined behavior.
	for match == match&mask[maskIdx] {
		// We don't store the match value in the node.
		// Bit 1 goes to right child, bit 0 goes to left child.
		if match == match&key[maskIdx] {
			next = node.right
		} else {
			next = node.left
		}

		// If we can't go further, break
		if nil == next {
			break
		}

		node = next

		if match == 1 {
			// Move to next byte in key/mask
			// If we have exhausted the key/mask, break
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			// Reset match to MSB
			match = msbByteVal
		} else {
			// Move to next bit in current byte
			match >>= 1
		}
	}

	// We found an existing node. This condition will evaluate to true
	// if we have covered the entire key/mask in the traversal above.
	if nil != next {
		// Cannot be hit but check just in case.
		// This cannot be the root node
		if t.IsRoot(node) {
			return Error, ErrInsertFailed
		}

		// If the node is already terminal, it's a duplicate insert
		// It is left to the caller to determine if this is an error.
		// We will not return an error here.
		if node.IsTerminal() {
			return Dup, nil
		}

		// Mark the node as terminal and set the value
		node.SaveAndMarkTerminal(value)

		// Increment node count
		t.incrNumNodes()

		// Successful insert
		return Ok, nil
	}

	// We are unlikely to hit this condition. Check for safety (future proofing).
	// The for loop above might change and make this condition evaluate to true.
	if keyLen == maskIdx {
		return Error, ErrInsertFailed
	}

	// Create new nodes for the remaining bits in the key/mask.
	for match == match&mask[maskIdx] {
		// Create a new node
		next = NewNode[T]()

		// Bit 1 goes to right child, bit 0 goes to left child.
		if match == match&key[maskIdx] {
			node.right = next
		} else {
			node.left = next
		}

		node = next

		if match == 1 {
			// Move to next byte in key/mask
			// If we have exhausted the key/mask, break
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			// Reset match to MSB
			match = msbByteVal
		} else {
			// Move to next bit in current byte
			match >>= 1
		}
	}

	// The last node created corresponds to the key/mask.
	// Mark it as terminal and set the value.
	node.SaveAndMarkTerminal(value)

	// Increment node count
	t.incrNumNodes()

	// Successful insert
	return Ok, nil
}

// find a key in the prefix tree. Caller must hold appropriate locks.
// Arguments:
//
//	key   - key to find expressed as byte slice.
//	mask  - mask for the key expressed as byte slice.
//	mType - type of match to perform (Exact/Partial)
//	nodeAncestors - stack of ancestor nodes. Optional argument.
//
// Returns:
//
//	*treeNode - pointer to the found node
//	*treeNodeStack - stack of nodes traversed during the search
//	OpResult  - result of the operation
//	error     - error if any
func (t *Tree[T]) find(key []byte, mask []byte, mType MatchType, nodeAncestors *NodeStack[T]) (*Node[T], OpResult, error) {
	if t.IsEmpty() {
		return nil, NoMatch, ErrKeyNotFound
	}

	keyLen := len(key)
	if keyLen <= 0 {
		return nil, Error, fmt.Errorf("invalid key length %d", keyLen)
	}

	match := msbByteVal
	maskIdx := 0

	// The very first bit of mask cannot be 0
	// The only way to store this node is to mark
	// the root as terminal. This is not supported.
	if match != match&mask[maskIdx] {
		return nil, Error, ErrInvalidKeyMask
	}

	// Start from root
	node := t.root.Node
	ret := Match

	// Traverse down the tree as far as possible.
	for nil != node && match == match&mask[maskIdx] {
		// Check for partial match condition. If we see a terminal node
		// during traversal and the match type is Partial, we are done.
		// A partial match will find the earliest matching prefix in the tree.
		if Partial == mType && node.IsTerminal() {
			ret = PartialMatch
			break
		}

		// Save the traversed node if asked for
		if nodeAncestors != nil {
			nodeAncestors.Push(node)
		}

		// Bit 1 goes to right child, bit 0 goes to left child.
		if match == match&key[maskIdx] {
			node = node.right
		} else {
			node = node.left
		}

		if match == 1 {
			// Move to next byte in key/mask
			// If we have exhausted the key/mask, break
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			// Reset match to MSB
			match = msbByteVal
		} else {
			// Move to next bit in current byte
			match >>= 1
		}
	}

	// For Exact match, we must end up on a terminal node
	if nil != node && node.IsTerminal() {
		return node, ret, nil
	}

	return nil, NoMatch, ErrKeyNotFound
}

// Delete a key from the prefix tree. Will write lock the tree when deleting.
// Arguments:
//
//	ctx  - context for the lock functions.
//	key  - key to delete expressed as byte slice.
//	mask - mask for the key expressed as byte slice.
//
// Returns:
//
//	OpResult - result of the operation
//	interface{} - value associated with the deleted key
//	error    - error if any
func (t *Tree[T]) Delete(ctx context.Context, key []byte, mask []byte) (OpResult, T, error) {
	var zero T
	if len(key) != len(mask) {
		return Error, zero, ErrInvalidKeyMask
	}

	t.wlock(ctx)
	defer func() {
		t.unlock(ctx)
	}()

	// Stack of ancestors to the node we are searching for
	nodeAncestors := NewNodeStack[T]()

	// Find the node to delete. It must be an exact match for deletion.
	node, result, err := t.find(key, mask, Exact, nodeAncestors)
	if nil != err || Match != result {
		return Error, zero, err
	}

	// This condition should never be hit
	if nil == node || !node.IsTerminal() || t.IsRoot(node) {
		return Error, zero, ErrKeyNotFound
	}

	// Is the match node not a leaf?
	if !node.IsLeaf() {
		// Unmark terminal to indicate deletion
		node.UnmarkTerminal()

		// Decrement node count
		t.decrNumNodes()

		// Deleted successfully
		return Match, node.value, nil
	}

	value := node.value

	// Remove nodes up the tree
	for !nodeAncestors.IsEmpty() {
		// Pop the parent node
		parent := nodeAncestors.Pop()

		// Remove the reference to the current node from the parent
		if node == parent.right {
			parent.right = nil
		} else {
			parent.left = nil
		}

		node = parent

		// If the new node is not a leaf, is a terminal or root, break
		if !node.IsLeaf() || node.IsTerminal() || t.IsRoot(node) {
			break
		}
	}

	// Decrement node count
	t.decrNumNodes()

	// Deleted successfully
	return Match, value, nil
}

// Searches for a key in the prefix tree. Will read lock the tree when searching.
// Arguments:
//
//	ctx   - context for the lock functions.
//	key   - key to find expressed as byte slice.
//	mask  - mask for the key expressed as byte slice.
//	mType - type of match to perform (Exact/Partial)
//
// Returns:
//
//	OpResult - result of the operation
//	interface{} - value associated with the found key
//	error    - error if any
func (t *Tree[T]) Search(ctx context.Context, key []byte, mask []byte, mType MatchType) (OpResult, T, error) {
	var zero T
	if len(key) != len(mask) {
		return Error, zero, ErrInvalidKeyMask
	}

	t.rlock(ctx)
	defer func() {
		t.runlock(ctx)
	}()

	// Find the node. Match type is determined by caller.
	node, result, err := t.find(key, mask, mType, nil)
	if nil != err {
		return Error, zero, err
	}

	// Validate result based on match type
	switch mType {
	case Exact:
		if result != Match {
			return Error, zero, err
		}

	case Partial:
		if result != Match && result != PartialMatch {
			return Error, zero, err
		}
	}

	// This condition should never be hit
	if nil == node || !node.IsTerminal() || t.IsRoot(node) {
		return Error, zero, ErrKeyNotFound
	}

	// Search successful
	return result, node.value, nil
}

// Searches for an exact match of the key in the prefix tree.
// Arguments:
//
//	ctx   - context for the lock functions.
//	key   - key to find expressed as byte slice.
//	mask  - mask for the key expressed as byte slice.
//
// Returns:
//
//	OpResult - result of the operation
//	interface{} - value associated with the found key
//	error    - error if any
func (t *Tree[T]) SearchExact(ctx context.Context, key []byte, mask []byte) (OpResult, T, error) {
	return t.Search(ctx, key, mask, Exact)
}

// Searches for a partial match of the key in the prefix tree.
// Arguments:
//
//	ctx   - context for the lock functions.
//	key   - key to find expressed as byte slice.
//	mask  - mask for the key expressed as byte slice.
//
// Returns:
//
//	OpResult - result of the operation
//	interface{} - value associated with the found key
//	error    - error if any
func (t *Tree[T]) SearchPartial(ctx context.Context, key []byte, mask []byte) (OpResult, T, error) {
	return t.Search(ctx, key, mask, Partial)
}

// Walk the tree using the provided walker function. Performs a depth-first traversal.
// The walker function is called for each node with a valid key and value.
// The k/v pairs are returned in the order they are encountered during the traversal.
// This might be different from the order in which they were inserted.
// Arguments:
//
//	ctx        - context for the lock functions.
//	walkerFn   - function to call for each node during the walk
//
// Returns:
//
//	error    - error if any
func (t *Tree[T]) Walk(ctx context.Context, walkerFn TreeWalkerFn[T]) error {
	if nil == walkerFn {
		return ErrNoWalkerFunction
	}

	if t.IsEmpty() {
		return nil
	}

	t.rlock(ctx)
	defer func() {
		t.runlock(ctx)
	}()

	// Node stack
	treeNodeStack := NewNodeStack[T]()

	// Start at root
	treeNodeStack.Push(t.root.Node)

	// Start looping
	for !treeNodeStack.IsEmpty() {
		// Peek at the top node in the stack
		node := treeNodeStack.Peek()

		// If there is a left child, push the left child onto the stack
		// and continue
		if nil != node.left {
			treeNodeStack.Push(node.left)
			continue
		}

		// If there is a right child, add the right child to the stack
		if nil != node.right {
			treeNodeStack.Push(node.right)
			continue
		}

		// Pop the current node from the stack
		node = treeNodeStack.Pop()

		// Ignore root
		if t.IsRoot(node) {
			continue
		}

		// Left node must be a terminal node. Call the walker function
		if node.IsTerminal() {
			err := walkerFn(ctx, node.value)
			if nil != err {
				return err
			}
		}

		// Unwind the stack to find the next unvisited right child
		for !treeNodeStack.IsEmpty() {
			parent := treeNodeStack.Peek()

			// If the parent has a right child and the right child
			// is not the current node. Push the right child onto
			// the stack and break.
			if nil != parent.right && parent.right != node {
				treeNodeStack.Push(parent.right)
				break
			}

			// Pop the parent node
			node = treeNodeStack.Pop()

			// If the popped node is not root node and is a terminal node, call the walker function
			if !t.IsRoot(node) && node.IsTerminal() {
				err := walkerFn(ctx, node.value)
				if nil != err {
					return err
				}
			}
		}
	}

	return nil
}
