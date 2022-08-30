package prefix_tree

import (
    "fmt"
    "net"
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
    v4MaskMsb [ net.IPv4len ]byte = [ net.IPv4len ]byte{ msbByteVal, 0, 0, 0 }
    v6MaskMsg [ net.IPv6len ]byte = [ net.IPv6len ]byte{ msbByteVal, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 }
)

func ( t *tree )Insertv4( saddr string, value interface{ } )( OpResult, error ) {
    if nil == t {
        return Err, fmt.Errorf( "invalid prefix tree" )
    }

    match   := v4MaskMsb
    maskIdx := 0

    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, err
    }

    t.wlock( )    
    defer func( ) {
        t.unlock( )
    }( )

    node := t.root
    next := t.root

    for 1 == match[ maskIdx ] & mask[ maskIdx ] {
        if 1 == addr[ maskIdx ] & match[ maskIdx ] {
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
            if net.IPv4len == maskIdx {
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

    if net.IPv4len == maskIdx {
        return Err, fmt.Errorf( "failed to add %s", saddr )
    }

    for 1 == addr[ maskIdx ] & mask[ maskIdx ] {
        next = newNode( )

        next.parent = node

        if 1 == addr[ maskIdx ] & match[ maskIdx ] {
            node.right = next
        } else {
            node.left = next
        }

        node = next

        if 1 == match[ maskIdx ] {
            maskIdx++
            if net.IPv4len == maskIdx {
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
func ( t *tree )findv4( addr net.IP, mask net.IPMask, mType MatchType )( *treeNode, OpResult, error ) {
    if nil == t {
        return nil, Err, fmt.Errorf( "invalid prefix tree" )
    }

    match   := v4MaskMsb
    maskIdx := 0

    node := t.root

    ret := Match

    for nil != node && 1 == match[ maskIdx ] & mask[ maskIdx ] {
        if Partial == mType && node.isTerminal( ) {
            ret = PartialMatch
            break
        }

        if 1 == addr[ maskIdx ] & match[ maskIdx ] {
            node = node.right
        } else {
            node = node.left
        }

        if 1 == match[ maskIdx ] {
            maskIdx++
            if net.IPv4len == maskIdx {
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

func ( t *tree )Deletev4( saddr string )( OpResult, interface{ }, error ) {
    if nil == t {
        return Err, nil, fmt.Errorf( "invalid prefix tree" )
    }

    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    t.wlock( )
    defer func( ) {
        t.unlock( )
    }( )

    node, result, err := t.findv4( addr, mask, Exact )
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

func ( t *tree )Searchv4( saddr string, mType MatchType )( OpResult, interface{ }, error ) {
    if nil == t {
        return Err, nil, fmt.Errorf( "invalid prefix tree" )
    }

    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    t.rlock( )
    defer func( ) {
        t.runlock( )
    }( )

    node, result, err := t.findv4( addr, mask, mType )
    if nil != err || Match != result {
        return Err, nil, err
    }

    // This condition should never be hit
    if nil == node || !node.isTerminal( ) || node.isRoot( ) {
        return Err, nil, fmt.Errorf( "node not found" )
    }

    return Match, node.value, nil
}
