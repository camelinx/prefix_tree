package prefix_tree

import (
	"context"
	"fmt"
	"net"
)

type V4Tree struct {
	tree *Tree
}

func testgetv4Addr(saddr string) (net.IP, net.IPMask, error) {
	return getv4Addr(saddr)
}

// Returns the IPv4 address and mask for the given string representation
// Arguments:
//
//	saddr - string representation of the IPv4 address. Can be in
//	         CIDR notation or just the IP address.
//
// Returns:
//
//	net.IP     - IPv4 address
//	net.IPMask - IPv4 mask
//	error      - error, if any
func getv4Addr(saddr string) (net.IP, net.IPMask, error) {
	// Try CIDR notation parsing first
	_, ipnet, err := net.ParseCIDR(saddr)
	if nil == err {
		// Ensure it's an IPv4 address
		if nil == ipnet.IP.To4() {
			return nil, nil, fmt.Errorf("invalid v4 address %s", saddr)
		}

		// Return the IPv4 address and mask
		return ipnet.IP, ipnet.Mask, nil
	}

	// Try parsing as a plain IPv4 address
	nip := net.ParseIP(saddr)
	if nil != nip && nil != nip.To4() {
		// Return the IPv4 address with a /32 mask
		return nip, net.CIDRMask(32, 32), nil
	}

	return nil, nil, fmt.Errorf("invalid v4 address %s", saddr)
}

// Returns a new IPv4 prefix tree
// Returns:
//
//	AddrTree - IPv4 prefix tree
func NewV4Tree() AddrTree {
	return &V4Tree{
		tree: NewTree(),
	}
}

// Sets the lock handlers for the IPv4 prefix tree
// Arguments:
//
//	rlockFn   - read lock function
//	runlockFn - read unlock function
//	wlockFn   - write lock function
//	unlockFn  - unlock function
func (v4t *V4Tree) SetLockHandlers(rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) {
	v4t.tree.SetLockHandlers(rlockFn, runlockFn, wlockFn, unlockFn)
}

// Inserts the given IPv4 address and mask into the tree
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv4 address. Can be in
//		    CIDR notation or just the IP address.
//	value - Optional value to be associated with the given address/mask. Can be nil.
//
// Returns:
//
//	OpResult - result of the insert operation
//	error    - error, if any
func (v4t *V4Tree) Insert(ctx context.Context, saddr string, value interface{}) (OpResult, error) {
	addr, mask, err := getv4Addr(saddr)
	if nil != err {
		return Error, err
	}

	// Insert parsed address and mask into the tree
	return v4t.tree.Insert(ctx, addr.To4(), mask, value)
}

// Deletes the given IPv4 address and mask from the tree
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv4 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the delete operation
//	interface{} - value associated with the deleted address/mask, if any
//	error    - error, if any
func (v4t *V4Tree) Delete(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv4Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	// Delete parsed address and mask from the tree
	return v4t.tree.Delete(ctx, addr.To4(), mask)
}

// Searches for the given IPv4 address and mask in the tree.
// Performs a partial search. If there is a prefix in the tree
// that matches the given address/mask, it will be returned.
// For e.g. if the tree has 192.168.128.0/24 and the search
// is for 192.168.128.40/32, the search will be successful.
// For exact match searches, use SearchExact().
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv4 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the search operation
//	interface{} - value associated with the found address/mask, if any
//	error    - error, if any
func (v4t *V4Tree) Search(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv4Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	// Perform partial search for parsed address and mask in the tree
	return v4t.tree.SearchPartial(ctx, addr.To4(), mask)
}

// Similar to Search(), but performs an exact match search.
// Arguments:
//
//	ctx   - context for the operation
//	saddr - string representation of the IPv4 address. Can be in
//		    CIDR notation or just the IP address.
//
// Returns:
//
//	OpResult - result of the search operation
//	interface{} - value associated with the found address/mask, if any
//	error    - error, if any
func (v4t *V4Tree) SearchExact(ctx context.Context, saddr string) (OpResult, interface{}, error) {
	addr, mask, err := getv4Addr(saddr)
	if nil != err {
		return Error, nil, err
	}

	return v4t.tree.SearchExact(ctx, addr.To4(), mask)
}

// Returns the number of nodes in the IPv4 prefix tree
// Returns:
//
//	uint64 - number of nodes in the tree
func (v4t *V4Tree) GetNodesCount() uint64 {
	return v4t.tree.NumNodes
}
