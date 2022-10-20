package prefix_tree

import (
    "fmt"
    "net"
)

type V6Tree struct {
    tree   *Tree
}

func ( v6t *V6Tree )testgetv6Addr( saddr string )( net.IP, net.IPMask, error ) {
    return getv6Addr( saddr )
}

func getv6Addr( saddr string )( net.IP, net.IPMask, error ) {
    _, ipnet, err := net.ParseCIDR( saddr )
    if nil == err {
        if nil == ipnet.IP.To16( ) || nil != ipnet.IP.To4( ) {
            return nil, nil, fmt.Errorf( "invalid v6 address %s", saddr )
        }

        return ipnet.IP, ipnet.Mask, nil
    }

    nip := net.ParseIP( saddr )
    if nil != nip && nil != nip.To16( ) && nil == nip.To4( ) {
        return nip, net.CIDRMask( 128, 128 ), nil
    }

    return nil, nil, fmt.Errorf( "invalid v6 address %s", saddr )
}

func NewV6Tree( )( *V6Tree ) {
    return &V6Tree{
        tree:   NewTree( ),
    }
}

func ( v6t *V6Tree )SetLockHandlers( lockCtx interface{ }, rlockFn ReadLockFn, runlockFn ReadUnlockFn, wlockFn WriteLockFn, unlockFn UnlockFn )( ) {
    v6t.tree.SetLockHandlers( lockCtx, rlockFn, runlockFn, wlockFn, unlockFn )
}

func ( v6t *V6Tree )Insert( saddr string, value interface{ } )( OpResult, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, err
    }

    return v6t.tree.Insert( addr, mask, net.IPv6len, value )
}

func ( v6t *V6Tree )Delete( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v6t.tree.Delete( addr, mask, net.IPv6len )
}

func ( v6t *V6Tree )Search( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v6t.tree.SearchPartial( addr, mask, net.IPv6len )
}

func ( v6t *V6Tree )SearchExact( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return v6t.tree.SearchExact( addr, mask, net.IPv6len )
}
