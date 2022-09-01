package prefix_tree

import (
    "fmt"
)

type ReadLockFn func( interface{ } )( )
type ReadUnlockFn func( interface{ } )( )
type WriteLockFn func( interface{ } )( )
type UnlockFn func( interface{ } )( )

type tree struct {
    root        *treeNode

    Nodes        uint64

    lockCtx      interface{ }
    rlockFn      ReadLockFn
    runlockFn    ReadUnlockFn
    wlockFn      WriteLockFn
    unlockFn     UnlockFn
}

func Init( )( *tree ) {
    return &tree{
        root: &treeNode{
            parent: nil,
        },

        lockCtx:   nil,
    }
}

func ( t *tree )SetLockHandlers( lockCtx interface{ }, rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn )( ) {
    if nil != t {
        t.lockCtx   = lockCtx
        t.rlockFn   = rlockFn
        t.runlockFn = runlockFn
        t.wlockFn   = wlockFn
        t.unlockFn  = unlockFn
    }
}

func ( t *tree )rlock( )( ) {
    if nil == t || nil == t.rlockFn {
        return
    }

    t.rlockFn( t.lockCtx )
}

func ( t *tree )runlock( )( ) {
    if nil == t || nil == t.runlockFn {
        return
    }

    t.runlockFn( t.lockCtx )
}

func ( t *tree )wlock( )( ) {
    if nil == t || nil == t.wlockFn {
        return
    }

    t.wlockFn( t.lockCtx )
}

func ( t *tree )unlock( )( ) {
    if nil == t || nil == t.unlockFn {
        return
    }

    t.unlockFn( t.lockCtx )
}

type OpResult int

const (
    Err   OpResult = iota
    Ok
    Dup
    Match
    PartialMatch
    NoMatch
)

type MatchType int

const (
    Exact   MatchType = iota
    Partial
)

var (
    msbByteVal byte = byte( 128 )
)

func ( t *tree )Insert( key [ ]byte, mask [ ]byte, keyLen int, value interface{ } )( OpResult, error ) {
    if nil == t {
        return Err, fmt.Errorf( "invalid prefix tree" )
    }

    if keyLen <= 0 {
        return Err, fmt.Errorf( "invalid key length %d", keyLen )
    }

    match   := make( [ ]byte, keyLen )
    maskIdx := 0

    match[ maskIdx ] = msbByteVal

    t.wlock( )
    defer func( ) {
        t.unlock( )
    }( )

    node := t.root
    next := t.root

    for 1 == match[ maskIdx ] & mask[ maskIdx ] {
        if 1 == key[ maskIdx ] & match[ maskIdx ] {
            next = node.right
        } else {
            next = node.left
        }

        if nil == next {
            break
        }

        node = next

        if 1 == match[ maskIdx ] {
            maskIdx++
            if keyLen == maskIdx {
                break
            }

            match[ maskIdx ] = msbByteVal
        } else {
            match[ maskIdx ] >>= 1
        }
    }

    if nil != next {
        if node.isTerminal( ) {
            return Dup, nil
        }

        node.value = value
        node.markTerminal( )

        return Ok, nil
    }

    if keyLen == maskIdx {
        return Err, fmt.Errorf( "insert failed" )
    }

    for 1 == key[ maskIdx ] & mask[ maskIdx ] {
        next = newNode( )

        next.parent = node

        if 1 == key[ maskIdx ] & match[ maskIdx ] {
            node.right = next
        } else {
            node.left = next
        }

        node = next

        if 1 == match[ maskIdx ] {
            maskIdx++
            if keyLen == maskIdx {
                break
            }

            match[ maskIdx ] = msbByteVal
        } else {
            match[ maskIdx ] >>= 1
        }
    }

    node.value = value
    node.markTerminal( )

    return Ok, nil
}

// Caller must lock
func ( t *tree )find( key [ ]byte, mask [ ]byte, keyLen int, mType MatchType )( *treeNode, OpResult, error ) {
    if nil == t {
        return nil, Err, fmt.Errorf( "invalid prefix tree" )
    }

    if keyLen <= 0 {
        return nil, Err, fmt.Errorf( "invalid key length %d", keyLen )
    }

    match   := make( [ ]byte, keyLen )
    maskIdx := 0

    match[ maskIdx ] = msbByteVal

    node := t.root
    ret  := Match

    for nil != node && 1 == match[ maskIdx ] & mask[ maskIdx ] {
        if Partial == mType && node.isTerminal( ) {
            ret = PartialMatch
            break
        }

        if 1 == key[ maskIdx ] & match[ maskIdx ] {
            node = node.right
        } else {
            node = node.left
        }

        if 1 == match[ maskIdx ] {
            maskIdx++
            if keyLen == maskIdx {
                break
            }

            match[ maskIdx ] = msbByteVal
        } else {
            match[ maskIdx ] >>= 1
        }
    }

    if nil != node && node.isTerminal( ) {
        return node, ret, nil
    }

    return nil, NoMatch, fmt.Errorf( "not found" )
}

func ( t *tree )Delete( key [ ]byte, mask [ ]byte, keyLen int )( OpResult, interface{ }, error ) {
    if nil == t {
        return Err, nil, fmt.Errorf( "invalid prefix tree" )
    }

    t.wlock( )
    defer func( ) {
        t.unlock( )
    }( )

    node, result, err := t.find( key, mask, keyLen, Exact )
    if nil != err || Match != result {
        return Err, nil, err
    }

    // This condition should never be hit
    if nil == node || !node.isTerminal( ) || node.isRoot( ) {
        return Err, nil, fmt.Errorf( "node not found" )
    }

    if !node.isLeaf( ) {
        node.unmarkTerminal( )
        return Match, node.value, nil
    }

    for true {
        if node == node.parent.right {
            node.parent.right = nil
        } else {
            node.parent.left = nil
        }

        node = node.parent

        if !node.isLeaf( ) || node.isTerminal( ) || node.isRoot( ) {
            break
        }
    }

    return Match, node.value, nil
}

func ( t *tree )Search( key [ ]byte, mask [ ]byte, keyLen int, mType MatchType )( OpResult, interface{ }, error ) {
    if nil == t {
        return Err, nil, fmt.Errorf( "invalid prefix tree" )
    }

    t.rlock( )
    defer func( ) {
        t.runlock( )
    }( )

    node, result, err := t.find( key, mask, keyLen, mType )
    if nil != err || Match != result {
        return Err, nil, err
    }

    // This condition should never be hit
    if nil == node || !node.isTerminal( ) || node.isRoot( ) {
        return Err, nil, fmt.Errorf( "node not found" )
    }

    return Match, node.value, nil
}

func (t *tree )SearchExact( key [ ]byte, mask [ ]byte, keyLen int )( OpResult, interface{ }, error ) {
    return t.Search( key, mask, keyLen, Exact )
}

func (t *tree )SearchPartial( key [ ]byte, mask [ ]byte, keyLen int )( OpResult, interface{ }, error ) {
    return t.Search( key, mask, keyLen, Partial )
}
