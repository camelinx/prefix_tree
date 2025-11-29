package prefix_tree

// Wrapper around prefix tree to support store and lookup of strings. Storing and looking up strings
// does not require a mask. This thin wrapper will abstract the use of masks for the user.
// All exported functions will have a reverse equivalent to support use cases like domain names.

import (
	"context"
)

type ReversedStringsTree[T any] struct {
	stree PrefixTree[T]
}

// Returns a new IPv4 prefix tree
// Returns:
//
//	AddrTree - IPv4 prefix tree
func NewReversedStringsTree[T any]() PrefixTree[T] {
	return &ReversedStringsTree[T]{
		stree: NewStringsTree[T](),
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
func NewReversedStringsTreeWithLockHandlers[T any](rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) PrefixTree[T] {
	return &ReversedStringsTree[T]{
		stree: NewStringsTreeWithLockHandlers[T](rlockFn, runlockFn, wlockFn, unlockFn),
	}
}

// Reverses a string
// Arguments:
//
//	s - string to be reversed
//
// Returns:
//
//	string - reversed string
func reverseString(s string) string {
	// A rune slice is needed to properly handle multi-byte characters
	// Reversing a byte slice does not guarantee correct results for multi-byte characters
	sr := []rune(s)
	for i, j := 0, len(sr)-1; i < j; i, j = i+1, j-1 {
		sr[i], sr[j] = sr[j], sr[i]
	}

	return string(sr)
}

// Insert the reversed string into the tree
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
func (rst *ReversedStringsTree[T]) Insert(ctx context.Context, s string, value *T) (OpResult, error) {
	return rst.stree.Insert(ctx, reverseString(s), value)
}

// Deletes the reversed string from the tree
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
func (rst *ReversedStringsTree[T]) Delete(ctx context.Context, s string) (OpResult, *T, error) {
	return rst.stree.Delete(ctx, reverseString(s))
}

// Searches for the reversed string in the tree.
// Performs a partial search. If there is a prefix in the tree
// that matches the reversed string, it will be returned.
// For exact match searches, use SearchExactReversed().
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
func (rst *ReversedStringsTree[T]) Search(ctx context.Context, s string) (OpResult, *T, error) {
	return rst.stree.Search(ctx, reverseString(s))
}

// Similar to SearchReversed(), but performs an exact match search.
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
func (rst *ReversedStringsTree[T]) SearchExact(ctx context.Context, s string) (OpResult, *T, error) {
	return rst.stree.SearchExact(ctx, reverseString(s))
}

// Returns the number of nodes in the IPv4 prefix tree
// Returns:
//
//	uint64 - number of nodes in the tree
func (rst *ReversedStringsTree[T]) GetNodesCount() uint64 {
	return rst.stree.GetNodesCount()
}
