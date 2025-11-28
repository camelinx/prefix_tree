package prefix_tree

import (
	"context"
	"fmt"
)

type Tree struct {
	root *treeNode

	NumNodes uint64

	rlockFn   ReadLockFn
	runlockFn ReadUnlockFn
	wlockFn   WriteLockFn
	unlockFn  UnlockFn
}

func NewTree() *Tree {
	return &Tree{
		root:     rootNode(),
		NumNodes: 0,
	}
}

// Sets the lock handlers for the prefix tree
// Arguments:
//
//	lockCtx   - context to be passed to the lock/unlock functions
//	rlockFn   - read lock function
//	runlockFn - read unlock function
//	wlockFn   - write lock function
//	unlockFn  - unlock function
func (t *Tree) SetLockHandlers(rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) {
	if nil != t {
		t.rlockFn = rlockFn
		t.runlockFn = runlockFn
		t.wlockFn = wlockFn
		t.unlockFn = unlockFn
	}
}

func (t *Tree) rlock(ctx context.Context) {
	if nil == t || nil == t.rlockFn {
		return
	}

	t.rlockFn(ctx)
}

func (t *Tree) runlock(ctx context.Context) {
	if nil == t || nil == t.runlockFn {
		return
	}

	t.runlockFn(ctx)
}

func (t *Tree) wlock(ctx context.Context) {
	if nil == t || nil == t.wlockFn {
		return
	}

	t.wlockFn(ctx)
}

func (t *Tree) unlock(ctx context.Context) {
	if nil == t || nil == t.unlockFn {
		return
	}

	t.unlockFn(ctx)
}

func (t *Tree) incrNumNodes() {
	if nil != t {
		t.NumNodes++
	}
}

func (t *Tree) decrNumNodes() {
	if nil != t && t.NumNodes > 0 {
		t.NumNodes--
	}
}

var (
	msbByteVal byte = byte(0x80) // 1000 0000
)

// Insert a key into the prefix tree. Will write lock the tree when inserting.
// Arguments:
//
//	ctx  - context for the lock functions.
//	key  - key to insert expressed as byte slice.
//	mask - mask for the key expressed as byte slice. A mask with non-contiguous
//		   1s is considered unexpected and will lead to undefined behavior.
//	value - value associated with the key. This is optional and can be nil.
//
// Returns:
//
//	OpResult - result of the operation
//	error    - error if any
func (t *Tree) Insert(ctx context.Context, key []byte, mask []byte, value interface{}) (OpResult, error) {
	if nil == t {
		return Error, ErrInvalidPrefixTree
	}

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

	t.wlock(ctx)
	defer func() {
		t.unlock(ctx)
	}()

	// Start from root
	node := t.root
	next := t.root

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
		// If the node is already terminal, it's a duplicate insert
		// It is left to the caller to determine if this is an error.
		// We will not return an error here.
		if node.isTerminal() {
			return Dup, nil
		}

		// Mark the node as terminal and set the value
		node.saveAndMarkTerminal(value)

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
		next = newNode()

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
	node.saveAndMarkTerminal(value)

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
//
// Returns:
//
//		*treeNode - pointer to the found node
//	 *treeNodeStack - stack of nodes traversed during the search
//		OpResult  - result of the operation
//		error     - error if any
func (t *Tree) find(key []byte, mask []byte, mType MatchType) (*treeNode, *treeNodeStack, OpResult, error) {
	if nil == t {
		return nil, nil, Error, ErrInvalidPrefixTree
	}

	keyLen := len(key)
	if keyLen <= 0 {
		return nil, nil, Error, fmt.Errorf("invalid key length %d", keyLen)
	}

	match := msbByteVal
	maskIdx := 0

	// Start from root
	node := t.root
	ret := Match

	treeNodeStack := newTreeNodeStack()

	// Traverse down the tree as far as possible.
	for nil != node && match == match&mask[maskIdx] {
		// Check for partial match condition. If we see a terminal node
		// during traversal and the match type is Partial, we are done.
		// A partial match will find the earliest matching prefix in the tree.
		if Partial == mType && node.isTerminal() {
			ret = PartialMatch
			break
		}

		// Save the traversed node
		treeNodeStack.Push(node)

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
	if nil != node && node.isTerminal() {
		return node, treeNodeStack, ret, nil
	}

	return nil, nil, NoMatch, ErrKeyNotFound
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
func (t *Tree) Delete(ctx context.Context, key []byte, mask []byte) (OpResult, interface{}, error) {
	if nil == t {
		return Error, nil, ErrInvalidPrefixTree
	}

	if len(key) != len(mask) {
		return Error, nil, ErrInvalidKeyMask
	}

	t.wlock(ctx)
	defer func() {
		t.unlock(ctx)
	}()

	// Find the node to delete. It must be an exact match for deletion.
	node, nodeAncestors, result, err := t.find(key, mask, Exact)
	if nil != err || Match != result {
		return Error, nil, err
	}

	// This condition should never be hit
	if nil == node || !node.isTerminal() || node.isRoot() {
		return Error, nil, ErrKeyNotFound
	}

	// Is the match node not a leaf?
	if !node.isLeaf() {
		// Unmark terminal to indicate deletion
		node.unmarkTerminal()

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

		// If the new node is a leaf, terminal or root, break
		if !node.isLeaf() || node.isTerminal() || node.isRoot() {
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
func (t *Tree) Search(ctx context.Context, key []byte, mask []byte, mType MatchType) (OpResult, interface{}, error) {
	if nil == t {
		return Error, nil, ErrInvalidPrefixTree
	}

	if len(key) != len(mask) {
		return Error, nil, ErrInvalidKeyMask
	}

	t.rlock(ctx)
	defer func() {
		t.runlock(ctx)
	}()

	// Find the node. Match type is determined by caller.
	node, _, result, err := t.find(key, mask, mType)
	if nil != err || Match != result {
		return Error, nil, err
	}

	// This condition should never be hit
	if nil == node || !node.isTerminal() || node.isRoot() {
		return Error, nil, ErrKeyNotFound
	}

	// Search successful
	return Match, node.value, nil
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
func (t *Tree) SearchExact(ctx context.Context, key []byte, mask []byte) (OpResult, interface{}, error) {
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
func (t *Tree) SearchPartial(ctx context.Context, key []byte, mask []byte) (OpResult, interface{}, error) {
	return t.Search(ctx, key, mask, Partial)
}
