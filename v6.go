package prefix_tree

import (
    "fmt"
    "net"
)

type v6 struct {
    b bool
}

func ( v6 *v6 )testgetv6Addr( saddr string )( net.IP, net.IPMask, error ) {
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

func ( t *Tree )Insertv6( saddr string, value interface{ } )( OpResult, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, err
    }

    return t.Insert( addr, mask, net.IPv6len, value )
}

func ( t *Tree )Deletev6( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.Delete( addr, mask, net.IPv6len )
}

func ( t *Tree )Searchv6( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.SearchPartial( addr, mask, net.IPv6len )
}

func ( t *Tree )Searchv6Exact( saddr string )( OpResult, interface{ }, error ) {
    addr, mask, err := getv6Addr( saddr )
    if nil != err {
        return Err, nil, err
    }

    return t.SearchExact( addr, mask, net.IPv6len )
}
