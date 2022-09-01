package prefix_tree

import (
    "testing"
    "net"
    "fmt"
    "time"
    "math/rand"
)

const (
    v4GenCount = 256
)

func validatev4Addr( t *testing.T, saddr string, exip string, exmask string, negate bool )( ) {
    v4 := new( v4 )

    nip, nmask, err := v4.testgetv4Addr( saddr )
    if ( nil == err ) == negate {
        t.Fatalf( "getv4Addr: error validating %s and negate %v", saddr, negate )
    }

    if !negate && ( nip.String( ) != exip || exmask != nmask.String( ) ) {
        t.Fatalf( "getv4Addr: invalid values when validating %s and negate %v", saddr, negate )
    }
}

func TestV4( t *testing.T )( ) {
    rand.Seed( time.Now( ).UnixNano( ) )

    validatev4Addr( t, "", "", "", true )
    validatev4Addr( t, "192.168.128.40/33", "", "", true )
    validatev4Addr( t, "256.168.128.40/32", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1/64", "", "", true )

    tree := Init( )

    for class := ipv4AddrClassMin + 1; class < ipv4AddrClassMax; class++ {
        v4gen := newIpv4Generator( )

        err := v4gen.initIpv4Block( v4GenCount, class )
        if err != nil {
            t.Fatalf( "failed to initialize v4 ip addresses for class %v", class )
        }

        for i := 0; i < v4gen.count; i++ {
            addr := v4gen.block[ i ]

            validatev4Addr( t, addr, addr, net.CIDRMask( 32, 32 ).String( ), false )

            for m := 32; m > 0; m-- {
                cidr := addr + "/" + fmt.Sprint( m )

                _, netaddr, err := net.ParseCIDR( cidr )
                if err != nil {
                    t.Fatalf( "failed to parse %s, %v", cidr, err )
                }

                validatev4Addr( t, cidr, netaddr.IP.String( ), netaddr.Mask.String( ), false )
            }
        }

        for i := 0; i < v4gen.count; i++ {
            res, err := tree.Insertv4( v4gen.block[ i ], nil )
            if err != nil || res != Ok {
                t.Fatalf( "Failed to insert %s", v4gen.block[ i ] )
            }

            res, _, err = tree.Searchv4( v4gen.block[ i ] )
            if err != nil || res != Match {
                t.Fatalf( "Failed to find %s", v4gen.block[ i ] )
            }

            res, _, err = tree.Searchv4Exact( v4gen.block[ i ] + "/32" )
            if err != nil || res != Match {
                t.Fatalf( "Failed to find (exact) %s", v4gen.block[ i ] + "/32" )
            }

            res, err = tree.Insertv4( v4gen.block[ i ] + "/32", nil )
            if err != nil || res != Dup {
                t.Fatalf( "Failed to recognize %s as duplicate", v4gen.block[ i ] + "/32" )
            }

            res, err = tree.Insertv4( v4gen.block[ i ], nil )
            if err != nil || res != Dup {
                t.Fatalf( "Failed to recognize %s as duplicate", v4gen.block[ i ] )
            }

            res, _, err = tree.Deletev4( v4gen.block[ i ] + "/32" )
            if err != nil || res != Match {
                t.Fatalf( "Failed to delete %s", v4gen.block[ i ] + "/32" )
            }

            res, _, err = tree.Searchv4( v4gen.block[ i ] + "/32" )
            if nil == err || res != Err {
                t.Fatalf( "Found non existent key %s", v4gen.block[ i ] + "/32" )
            }

            res, _, err = tree.Searchv4Exact( v4gen.block[ i ] )
            if nil == err || res != Err {
                t.Fatalf( "Found (exact) non existent key %s", v4gen.block[ i ] )
            }

            res, _, err = tree.Deletev4( v4gen.block[ i ] )
            if nil == err || res != Err {
                t.Fatalf( "Deleted non existent key %s", v4gen.block[ i ] )
            }
        }
    }
}
