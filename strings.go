package prefix_tree

// Wrapper around prefix tree to support store and lookup of strings. Storing and looking up strings
// does not require a mask. This thin wrapper will abstract the use of masks for the user.
// All exported functions will have a reverse equivalent to support use cases like domain names.

import (
	"context"
)

type StringsTree[T any] struct {
	tree *Tree[T]
}

// Returns a new IPv4 prefix tree
// Returns:
//
//	AddrTree - IPv4 prefix tree
func NewStringsTree[T any]() PrefixTree[T] {
	return &StringsTree[T]{
		tree: NewTree[T](),
	}
}

// Returns a new IPv4 prefix tree with custom lock handlers
// Arguments:
//
//	rlockFn   - read lock function
//	runlockFn - read unlock function
//	wlockFn   - write lock function
//	unlockFn  - unlock function
//
// Returns:
//
//	AddrTree - IPv4 prefix tree
func NewStringsTreeWithLockHandlers[T any](rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) PrefixTree[T] {
	return &StringsTree[T]{
		tree: NewTreeWithLockHandlers[T](rlockFn, runlockFn, wlockFn, unlockFn),
	}
}

// Converts a string byte slice to a mask byte slice
// Arguments:
//
//	sb - string byte slice
//
// Returns:
//
//	[]byte - mask byte slice with all bits set
func getMaskFromString(sb []byte) []byte {
	mask := make([]byte, len(sb))
	for i := range mask {
		mask[i] = 0xFF
	}

	return mask
}

// Inserts the given string into the tree
// Arguments:
//
//	ctx   - context for the operation
//	s     - key as a string
//	value - Optional value to be associated with the given string. Can be nil.
//
// Returns:
//
//	OpResult - result of the insert operation
//	error    - error, if any
func (st *StringsTree[T]) Insert(ctx context.Context, s string, value T) (OpResult, error) {
	sb := []byte(s)

	// Insert parsed address and mask into the tree
	return st.tree.Insert(ctx, sb, getMaskFromString(sb), value)
}

// Deletes the given string from the tree
// Arguments:
//
//	ctx - context for the operation
//	s   - key as a string
//
// Returns:
//
//	OpResult - result of the delete operation
//	T        - value associated with the deleted address/mask, if any
//	error    - error, if any
func (st *StringsTree[T]) Delete(ctx context.Context, s string) (OpResult, T, error) {
	sb := []byte(s)

	// Delete parsed address and mask from the tree
	return st.tree.Delete(ctx, sb, getMaskFromString(sb))
}

// Searches for the given string in the tree.
// Performs a partial search. If there is a prefix in the tree
// that matches the given string, it will be returned.
// For exact match searches, use SearchExact().
// Arguments:
//
//	ctx - context for the operation
//	s   - key as a string
//
// Returns:
//
//	OpResult - result of the search operation
//	T        - value associated with the found address/mask, if any
//	error    - error, if any
func (st *StringsTree[T]) Search(ctx context.Context, s string) (OpResult, T, error) {
	sb := []byte(s)

	// Perform partial search for parsed address and mask in the tree
	return st.tree.SearchPartial(ctx, sb, getMaskFromString(sb))
}

// Similar to Search(), but performs an exact match search.
// Arguments:
//
//	ctx - context for the operation
//	s   - key as a string
//
// Returns:
//
//	OpResult - result of the search operation
//	T        - value associated with the found address/mask, if any
//	error    - error, if any
func (st *StringsTree[T]) SearchExact(ctx context.Context, s string) (OpResult, T, error) {
	sb := []byte(s)

	// Perform exact search for parsed address and mask in the tree
	return st.tree.SearchExact(ctx, sb, getMaskFromString(sb))
}

// Returns the number of nodes in the IPv4 prefix tree
// Returns:
//
//	uint64 - number of nodes in the tree
func (st *StringsTree[T]) GetNodesCount() uint64 {
	return st.tree.numNodes
}

// Walk the tree and call passed function for all nodes
// Arguments:
//
//	ctx - context for the operaton
//	callback - function to be called for every value in the tree
//
// Returns:
//
//	err - nil if successful else an error
func (st *StringsTree[T]) Walk(ctx context.Context, callback WalkerFn[T]) error {
	st.tree.Walk(ctx, func(ctx context.Context, value T) error {
		return callback(ctx, value)
	})

	return nil
}
