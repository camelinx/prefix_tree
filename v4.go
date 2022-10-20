package prefix_tree

import (
    "fmt"
    "net"
)

type V4Tree struct {
    tree   *Tree
}

func ( v4t *V4Tree )testgetv4Addr( saddr string )( net.IP, net.IPMask, error ) {
    return getv4Addr( saddr )
}

func getv4Addr( saddr string )( net.IP, net.IPMask, error ) {
    _, ipnet, err := net.ParseCIDR( saddr )
    if nil == err {
        if nil == ipnet.IP.To4( ) {
            return nil, nil, fmt.Errorf( "invalid v4 address %s", saddr )
        }

        return ipnet.IP, ipnet.Mask, nil
    }

    nip := net.ParseIP( saddr )
    if nil != nip && nil != nip.To4( ) {
        return nip, net.CIDRMask( 32, 32 ), nil
    }

    return nil, nil, fmt.Errorf( "invalid v4 address %s", saddr )
}

func NewV4Tree( )( *V4Tree ) {
    return &V4Tree{
        tree:   NewTree( ),
    }
}

func ( v4t *V4Tree )SetLockHandlers( lockCtx interface{ }, rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn )( ) {
    v4t.tree.SetLockHandlers( lockCtx, rlockFn, runlockFn, wlockFn, unlockFn )
}

func ( v4t *V4Tree )Insert( saddr string, value interface{ } )( OpResult, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, err
    }

    return v4t.tree.Insert( addr.To4( ), mask, net.IPv4len, value )
}

func ( v4t *V4Tree )Delete( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v4t.tree.Delete( addr.To4( ), mask, net.IPv4len )
}

func ( v4t *V4Tree )Search( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v4t.tree.SearchPartial( addr.To4( ), mask, net.IPv4len )
}

func ( v4t *V4Tree )SearchExact( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v4t.tree.SearchExact( addr.To4( ), mask, net.IPv4len )
}

func ( v4t *V4Tree )GetNodesCount( )( uint64 ) {
    return v4t.tree.NumNodes
}
