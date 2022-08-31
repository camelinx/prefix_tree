package prefix_tree

import (
    "fmt"
    "net"
)

type utils struct {
    b bool
}

func ( u *utils )testgetv4Addr( saddr string )( net.IP, net.IPMask, error ) {
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

func ( u *utils )testgetv6Addr( saddr string )( net.IP, net.IPMask, error ) {
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
