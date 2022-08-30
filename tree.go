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
    OpErr   OpResult = iota
    OpOk
    OpDup
)

var (
    msbByteVal byte = byte( 128 )
    v4MaskMsb [ net.IPv4len ]byte = [ net.IPv4len ]byte{ msbByteVal, 0, 0, 0 }
)

func ( t *tree )InsertV4( saddr string, value interface{ } )( OpResult, error ) {
    if nil == t {
        return OpErr, fmt.Errorf( "invalid prefix tree" )
    }

    match   := v4MaskMsb
    maskIdx := 0

    addr, mask, err := getV4Addr( saddr )
    if nil != err {
        return OpErr, err
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
            return OpDup, nil
        }

        node.value = value
        return OpOk, nil
    }

    if net.IPv4len == maskIdx {
        return OpErr, fmt.Errorf( "failed to add %s", saddr )
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
            match[ maskIdx ] = 128
        } else {
            match[ maskIdx ] >>= 1
        }
    }

    node.value = value
    return OpOk, nil
}
