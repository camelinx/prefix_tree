package prefix_tree

import (
    "testing"
    "net"
)

func validatev4Addr( t *testing.T, saddr string, exip string, exmask string, negate bool )( ) {
    u := new( utils )

    nip, nmask, err := u.testgetv4Addr( saddr )
    if ( nil == err ) == negate {
        t.Fatalf( "getv4Addr: error validating %s and negate %v", saddr, negate )
    }

    if !negate && ( nip.String( ) != exip || exmask != nmask.String( ) ) {
        t.Fatalf( "getv4Addr: invalid values when validating %s and negate %v", saddr, negate )
    }
}

func validatev6Addr( t *testing.T, saddr string, exip string, exmask string, negate bool )( ) {
    u := new( utils )

    nip, nmask, err := u.testgetv6Addr( saddr )
    if ( nil == err ) == negate {
        t.Fatalf( "getv6Addr: error validating %s and negate %v", saddr, negate )
    }

    if !negate && ( nip.String( ) != exip || exmask != nmask.String( ) ) {
        t.Fatalf( "getv6Addr: invalid values when validating %s and negate %v", saddr, negate )
    }
}

func TestUtils( t *testing.T )( ) {
    validatev4Addr( t, "192.168.128.40", "192.168.128.40", net.CIDRMask( 32, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/32", "192.168.128.40", net.CIDRMask( 32, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/24", "192.168.128.0", net.CIDRMask( 24, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/16", "192.168.0.0", net.CIDRMask( 16, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/12", "192.160.0.0", net.CIDRMask( 12, 32 ).String( ), false )
    validatev4Addr( t, "192.168.128.40/33", "", "", true )
    validatev4Addr( t, "256.168.128.40/32", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1", "", "", true )
    validatev4Addr( t, "2001:db8:a0b:12f0::1/64", "", "", true )

    validatev6Addr( t, "2001:db8:a0b:12f0::1", "2001:db8:a0b:12f0::1", net.CIDRMask( 128, 128 ).String( ), false )
    validatev6Addr( t, "2001:db8:a0b:12f0::1/128", "2001:db8:a0b:12f0::1", net.CIDRMask( 128, 128 ).String( ), false )
    validatev6Addr( t, "2001:db8:a0b:12f0::1/64", "2001:db8:a0b:12f0::", net.CIDRMask( 64, 128 ).String( ), false )
    validatev6Addr( t, "2001:db8:a0b:12f0::1/48", "2001:db8:a0b::", net.CIDRMask( 48, 128 ).String( ), false )
    validatev6Addr( t, "2001:db8:a0b:12f0::1/42", "2001:db8:a00::", net.CIDRMask( 42, 128 ).String( ), false )
    validatev6Addr( t, "59fb::1005:cc57:6571/128", "59fb::1005:cc57:6571", net.CIDRMask( 128, 128 ).String( ), false )
    validatev6Addr( t, "56fe::2159:5bbc::6594", "", "", true )
    validatev6Addr( t, "192.168.128.40", "", "", true )
    validatev6Addr( t, "192.168.128.40/32", "", "", true )
}
