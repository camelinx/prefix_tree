package prefix_tree

import (
    "fmt"
    "net"
)

type v4 struct {
    b bool
}

func ( v4 *v4 )testgetv4Addr( saddr string )( net.IP, net.IPMask, error ) {
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

func ( t *Tree )Insertv4( saddr string, value interface{ } )( OpResult, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, err
    }

    return t.Insert( addr.To4( ), mask, net.IPv4len, value )
}

func ( t *Tree )Deletev4( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.Delete( addr.To4( ), mask, net.IPv4len )
}

func ( t *Tree )Searchv4( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.SearchPartial( addr.To4( ), mask, net.IPv4len )
}

func ( t *Tree )Searchv4Exact( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv4Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.SearchExact( addr.To4( ), mask, net.IPv4len )
}
