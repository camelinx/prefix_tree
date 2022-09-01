package prefix_tree

import (
    "testing"
    "net"
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
    validatev4Addr( t, "", "", "", true )
    validatev4Addr( t, "192.168.128.40", "192.168.128.40", net.CIDRMask( 32, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/32", "192.168.128.40", net.CIDRMask( 32, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/24", "192.168.128.0", net.CIDRMask( 24, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/16", "192.168.0.0", net.CIDRMask( 16, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/12", "192.160.0.0", net.CIDRMask( 12, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/33", "", "", true )
    validatev4Addr( t, "256.168.128.40/32", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1/64", "", "", true )
}
