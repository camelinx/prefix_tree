package prefix_tree

import (
	"context"
	"errors"
)

type OpResult int

const (
	Error OpResult = iota
	Ok
	Dup
	Match
	PartialMatch
	NoMatch
)

type MatchType int

const (
	Exact MatchType = iota
	Partial
)

type ReadLockFn func(context.Context)
type ReadUnlockFn func(context.Context)
type WriteLockFn func(context.Context)
type UnlockFn func(context.Context)

type AddrTree interface {
	SetLockHandlers(ReadLockFn, ReadUnlockFn, WriteLockFn, UnlockFn)
	Insert(context.Context, string, interface{}) (OpResult, error)
	Delete(context.Context, string) (OpResult, interface{}, error)
	Search(context.Context, string) (OpResult, interface{}, error)
	SearchExact(context.Context, string) (OpResult, interface{}, error)
	GetNodesCount() uint64
}

var (
	ErrInvalidPrefixTree = errors.New("invalid prefix tree")
	ErrInvalidKeyMask    = errors.New("invalid key/mask")
	ErrInsertFailed      = errors.New("insert failed")
	ErrKeyNotFound       = errors.New("key not found")
)
