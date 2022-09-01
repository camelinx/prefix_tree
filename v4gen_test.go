package prefix_tree

import (
    "testing"
    "time"
    "math/rand"
)

const (
    ipv4GenMagicNum = 32
)

func testNewIpv4Generator( t *testing.T )( ipv4Gen *ipv4Gen ) {
    ipv4Gen = newIpv4Generator( )
    if nil == ipv4Gen {
        t.Fatalf( "newIpv4Generator - failed to initialize" )
    }

    return ipv4Gen
}

func testInitIpv4Block( t *testing.T ) {
    ipv4Gen := testNewIpv4Generator( t )
    err     := ipv4Gen.initIpv4Block( ipv4GenMagicNum, ipv4AddrClassAny )
    if err != nil || !ipv4Gen.initialized {
        t.Fatalf( "initIpv4Block - failed to initialize from count" )
    }

    ipv4Gen = testNewIpv4Generator( t )
    err     = ipv4Gen.initIpv4Block( ipv4GenMagicNum, ipv4AddrClassMin )
    if err == nil {
        t.Fatalf( "initIpv4Block - successfully initialized for invalid address class - lower bound" )
    }

    err = ipv4Gen.initIpv4Block( ipv4GenMagicNum, ipv4AddrClassMax )
    if err == nil {
        t.Fatalf( "initIpv4Block - successfully initialized for invalid address class - upper bound" )
    }

    for class := ipv4AddrClassAny; class <= ipv4AddrClassLoopback; class++ {
        ipv4Gen  = testNewIpv4Generator( t )
        err      = ipv4Gen.initIpv4Block( ipv4GenMagicNum, class )
        if err != nil || !ipv4Gen.initialized {
            t.Fatalf( "initIpv4Block - failed to initialize from count" )
        }

        for j := 0; j < ipv4GenMagicNum; j++ {
            err = ipv4Gen.validateIpv4Address( ipv4Gen.block[ j ] )
            if err != nil {
                t.Fatalf( "getIpv4Block - invalid ip address %v for count %v and class %v: error %v", ipv4Gen.block[ j ], ipv4GenMagicNum, class, err )
            }
        }
    }
}


type negativeTester func( *testing.T, *ipv4Gen  )

var negativeTesters = [ ]negativeTester {
    ipv4AddrClassAny        :   negativeTestClassAny,
    ipv4AddrClassA          :   negativeTestClassA,
    ipv4AddrClassAPrivate   :   negativeTestClassAPrivate,
    ipv4AddrClassLoopback   :   negativeTestClassLoopback,
}

func negativeTestClassAny( t *testing.T, ipv4Gen *ipv4Gen ) {
    invalidIps := [ ]string { "0.1.2.3", "11.12.13.256", "121.256.23.24", "131.32.256.34", "256.242.43.44", "256.256.256.256" }

    for _, invalidIp := range invalidIps {
        err := ipv4Gen.validateIpv4Address( invalidIp )
        if err == nil {
            t.Fatalf( "validateIpv4Address - failed to detect invalid ip address %v for class any", invalidIp )
        }
    }
}

func negativeTestClassA( t *testing.T, ipv4Gen *ipv4Gen ) {
    invalidIps := [ ]string { "0.1.2.3", "11.12.13.256", "21.256.23.24", "31.32.256.34", "256.42.43.44", "127.0.0.1", "10.0.0.1", "128.0.0.1", "192.0.0.1", "172.0.0.1" }

    for _, invalidIp := range invalidIps {
        err := ipv4Gen.validateIpv4Address( invalidIp )
        if err == nil {
            t.Fatalf( "validateIpv4Address - failed to detect invalid ip address %v for class A", invalidIp )
        }
    }
}

func negativeTestClassAPrivate( t *testing.T, ipv4Gen *ipv4Gen ) {
    invalidIps := [ ]string { "0.1.2.3", "11.12.13.256", "21.256.23.24", "31.32.256.34", "256.42.43.44", "127.0.0.1", "128.0.0.1", "192.0.0.1", "172.0.0.1" }

    for _, invalidIp := range invalidIps {
        err := ipv4Gen.validateIpv4Address( invalidIp )
        if err == nil {
            t.Fatalf( "validateIpv4Address - failed to detect invalid ip address %v for class A private", invalidIp )
        }
    }
}

func negativeTestClassLoopback( t *testing.T, ipv4Gen *ipv4Gen ) {
    invalidIps := [ ]string { "0.1.2.3", "11.12.13.256", "21.256.23.24", "31.32.256.34", "256.42.43.44", "10.0.0.1", "128.0.0.1", "192.0.0.1", "172.0.0.1" }

    for _, invalidIp := range invalidIps {
        err := ipv4Gen.validateIpv4Address( invalidIp )
        if err == nil {
            t.Fatalf( "validateIpv4Address - failed to detect invalid ip address %v for loopback", invalidIp )
        }
    }
}

func testValidateIpv4Address( t *testing.T ) {
    for class := ipv4AddrClassAny; class <= ipv4AddrClassLoopback; class++ {
        ipv4Gen := testNewIpv4Generator( t )
        err     := ipv4Gen.initIpv4Block( ipv4GenMagicNum, class )
        if err != nil || !ipv4Gen.initialized {
            t.Fatalf( "initIpv4Block - failed to initialize from count" )
        }

        for j := 0; j < ipv4GenMagicNum; j++ {
            err = ipv4Gen.validateIpv4Address( ipv4Gen.block[ j ] )
            if err != nil {
                t.Fatalf( "validateIpv4Address - invalid ip address %v for count %v and class %v: error %v", ipv4Gen.block[ j ], ipv4GenMagicNum, class, err )
            }
        }

        negativeTesters[ class ]( t, ipv4Gen )
    }
}

func TestIpv4Gen( t *testing.T ) {
    rand.Seed( time.Now( ).UnixNano( ) )
    testInitIpv4Block( t )
    testValidateIpv4Address( t )
}
