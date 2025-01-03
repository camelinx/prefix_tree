package prefix_tree

import (
	"fmt"
)

type Tree struct {
	root *treeNode

	NumNodes uint64

	lockCtx   interface{}
	rlockFn   ReadLockFn
	runlockFn ReadUnlockFn
	wlockFn   WriteLockFn
	unlockFn  UnlockFn
}

func NewTree() *Tree {
	return &Tree{
		root: &treeNode{
			parent: nil,
		},

		lockCtx: nil,
	}
}

func (t *Tree) SetLockHandlers(lockCtx interface{}, rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn) {
	if nil != t {
		t.lockCtx = lockCtx
		t.rlockFn = rlockFn
		t.runlockFn = runlockFn
		t.wlockFn = wlockFn
		t.unlockFn = unlockFn
	}
}

func (t *Tree) rlock() {
	if nil == t || nil == t.rlockFn {
		return
	}

	t.rlockFn(t.lockCtx)
}

func (t *Tree) runlock() {
	if nil == t || nil == t.runlockFn {
		return
	}

	t.runlockFn(t.lockCtx)
}

func (t *Tree) wlock() {
	if nil == t || nil == t.wlockFn {
		return
	}

	t.wlockFn(t.lockCtx)
}

func (t *Tree) unlock() {
	if nil == t || nil == t.unlockFn {
		return
	}

	t.unlockFn(t.lockCtx)
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
	msbByteVal byte = byte(128)
)

func (t *Tree) Insert(key []byte, mask []byte, keyLen int, value interface{}) (OpResult, error) {
	if nil == t {
		return Err, fmt.Errorf("invalid prefix tree")
	}

	if keyLen <= 0 {
		return Err, fmt.Errorf("invalid key length %d", keyLen)
	}

	match := make([]byte, keyLen)
	maskIdx := 0

	match[maskIdx] = msbByteVal

	t.wlock()
	defer func() {
		t.unlock()
	}()

	node := t.root
	next := t.root

	for match[maskIdx] == match[maskIdx]&mask[maskIdx] {
		if match[maskIdx] == match[maskIdx]&key[maskIdx] {
			next = node.right
		} else {
			next = node.left
		}

		if nil == next {
			break
		}

		node = next

		if match[maskIdx] == 1 {
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			match[maskIdx] = msbByteVal
		} else {
			match[maskIdx] >>= 1
		}
	}

	if nil != next {
		if node.isTerminal() {
			return Dup, nil
		}

		node.value = value
		node.markTerminal()

		t.incrNumNodes()

		return Ok, nil
	}

	if keyLen == maskIdx {
		return Err, fmt.Errorf("insert failed")
	}

	for match[maskIdx] == match[maskIdx]&mask[maskIdx] {
		next = newNode()

		next.parent = node

		if match[maskIdx] == match[maskIdx]&key[maskIdx] {
			node.right = next
		} else {
			node.left = next
		}

		node = next

		if match[maskIdx] == 1 {
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			match[maskIdx] = msbByteVal
		} else {
			match[maskIdx] >>= 1
		}
	}

	node.value = value
	node.markTerminal()

	t.incrNumNodes()

	return Ok, nil
}

// Caller must lock
func (t *Tree) find(key []byte, mask []byte, keyLen int, mType MatchType) (*treeNode, OpResult, error) {
	if nil == t {
		return nil, Err, fmt.Errorf("invalid prefix tree")
	}

	if keyLen <= 0 {
		return nil, Err, fmt.Errorf("invalid key length %d", keyLen)
	}

	match := make([]byte, keyLen)
	maskIdx := 0

	match[maskIdx] = msbByteVal

	node := t.root
	ret := Match

	for nil != node && match[maskIdx] == match[maskIdx]&mask[maskIdx] {
		if Partial == mType && node.isTerminal() {
			ret = PartialMatch
			break
		}

		if match[maskIdx] == match[maskIdx]&key[maskIdx] {
			node = node.right
		} else {
			node = node.left
		}

		if match[maskIdx] == 1 {
			maskIdx++
			if keyLen == maskIdx {
				break
			}

			match[maskIdx] = msbByteVal
		} else {
			match[maskIdx] >>= 1
		}
	}

	if nil != node && node.isTerminal() {
		return node, ret, nil
	}

	return nil, NoMatch, fmt.Errorf("not found")
}

func (t *Tree) Delete(key []byte, mask []byte, keyLen int) (OpResult, interface{}, error) {
	if nil == t {
		return Err, nil, fmt.Errorf("invalid prefix tree")
	}

	t.wlock()
	defer func() {
		t.unlock()
	}()

	node, result, err := t.find(key, mask, keyLen, Exact)
	if nil != err || Match != result {
		return Err, nil, err
	}

	// This condition should never be hit
	if nil == node || !node.isTerminal() || node.isRoot() {
		return Err, nil, fmt.Errorf("node not found")
	}

	if !node.isLeaf() {
		node.unmarkTerminal()

		t.decrNumNodes()

		return Match, node.value, nil
	}

	value := node.value

	for {
		if node == node.parent.right {
			node.parent.right = nil
		} else {
			node.parent.left = nil
		}

		node = node.parent

		if !node.isLeaf() || node.isTerminal() || node.isRoot() {
			break
		}
	}

	t.decrNumNodes()

	return Match, value, nil
}

func (t *Tree) Search(key []byte, mask []byte, keyLen int, mType MatchType) (OpResult, interface{}, error) {
	if nil == t {
		return Err, nil, fmt.Errorf("invalid prefix tree")
	}

	t.rlock()
	defer func() {
		t.runlock()
	}()

	node, result, err := t.find(key, mask, keyLen, mType)
	if nil != err || Match != result {
		return Err, nil, err
	}

	// This condition should never be hit
	if nil == node || !node.isTerminal() || node.isRoot() {
		return Err, nil, fmt.Errorf("node not found")
	}

	return Match, node.value, nil
}

func (t *Tree) SearchExact(key []byte, mask []byte, keyLen int) (OpResult, interface{}, error) {
	return t.Search(key, mask, keyLen, Exact)
}

func (t *Tree) SearchPartial(key []byte, mask []byte, keyLen int) (OpResult, interface{}, error) {
	return t.Search(key, mask, keyLen, Partial)
}
