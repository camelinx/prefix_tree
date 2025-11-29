package prefix_tree

import (
	"context"
	"fmt"
	"net"
)

type V6Tree struct {
	tree *Tree
}

func testgetv6Addr(saddr string) (net.IP, net.IPMask, error) {
	return getv6Addr(saddr)
}

// Returns the IPv6 address and mask for the given string representation
// Arguments:
//
//	saddr - string representation of the IPv6 address. Can be in
//	         CIDR notation or just the IP address.
//
// Returns:
//
//	net.IP     - IPv6 address
//	net.IPMask - IPv6 mask
//	error      - error, if any
func getv6Addr(saddr string) (net.IP, net.IPMask, error) {
	// Try CIDR notation parsing first
	_, ipnet, err := net.ParseCIDR(saddr)
	if nil == err {
		// Ensure it's an IPv6 address
		if nil == ipnet.IP.To16() || nil != ipnet.IP.To4() {
			return nil, nil, fmt.Errorf("invalid v6 address %s", saddr)
		}

		// Return the IPv6 address and mask
		return ipnet.IP, ipnet.Mask, nil
	}

	// Try parsing as a plain IPv6 address
	nip := net.ParseIP(saddr)
	if nil != nip && nil != nip.To16() && nil == nip.To4() {
		// Return the IPv6 address with a /128 mask
		return nip, net.CIDRMask(128, 128), nil
	}

	return nil, nil, fmt.Errorf("invalid v6 address %s", saddr)
}

// Returns a new IPv6 prefix tree
// Returns:
//
//	AddrTree - IPv6 prefix tree
func NewV6Tree() AddrTree {
	return &V6Tree{
		tree: NewTree(),
	}
}

// Returns a new IPv6 prefix tree with custom lock handlers
// Arguments:
//
//	rlockFn   - read lock function
//	runlockFn - read unlock function
//	wlockFn   - write lock function
//	unlockFn  - unlock function
//
// Returns:
//
//	AddrTree - IPv6 prefix tree
func NewV6TreeWithLockHandlers(rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) AddrTree {
	return &V6Tree{
		tree: NewTreeWithLockHandlers(rlockFn, runlockFn, wlockFn, unlockFn),
	}
}

// Inserts a new IPv6 address into the prefix tree
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv6 address. Can be in
//		    CIDR notation or just the IP address.
//	value - value to be associated with the IPv6 address
//
// Returns:
//
//	OpResult - result of the insert operation
//	error    - error, if any
func (v6t *V6Tree) Insert(ctx context.Context, saddr string, value interface{}) (OpResult, error) {
	addr, mask, err := getv6Addr(saddr)
	if nil != err {
		return Error, err
	}

	// Insert into the underlying tree
	return v6t.tree.Insert(ctx, addr, mask, value)
}

// Deletes the given IPv6 address from the prefix tree
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv6 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the delete operation
//	interface{} - value associated with the deleted address, if any
//	error      - error, if any
func (v6t *V6Tree) Delete(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv6Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	// Delete from the underlying tree
	return v6t.tree.Delete(ctx, addr, mask)
}

// Searches for the given IPv6 address in the prefix tree.
// Performs a partial search. If there is a prefix in the tree
// that matches the given address, it is returned. For e.g., if
// the tree contains 2001:db8::/32 and the search is for
// 2001:db8:abcd:0012::1/64, the search will be successful
// For exact match searches, use SearchExact() instead.
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv6 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the search operation
//	interface{} - value associated with the found address, if any
//	error      - error, if any
func (v6t *V6Tree) Search(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv6Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	// Perform partial search in the underlying tree
	return v6t.tree.SearchPartial(ctx, addr, mask)
}

// Similar to Search(), but performs an exact match search.
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv6 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the search operation
//	interface{} - value associated with the found address, if any
//	error      - error, if any
func (v6t *V6Tree) SearchExact(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv6Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	return v6t.tree.SearchExact(ctx, addr, mask)
}

// Returns the number of nodes in the IPv6 prefix tree
// Returns:
//
//	uint64 - number of nodes in the tree
func (v6t *V6Tree) GetNodesCount() uint64 {
	return v6t.tree.NumNodes
}
