package prefix_tree

import (
	"fmt"
	"math/rand"
	"net"
)

const (
	ipv4MinOctet                = 0
	ipv4MaxOctet                = 255
	ipv4ClassAMaxOctet          = 126
	ipv4ClassAPrivateFirstOctet = 10
	ipv4LoopbackFirstOctet      = 127
)

type ipv4AddrClass int

const (
	ipv4AddrClassMin ipv4AddrClass = iota
	ipv4AddrClassAny
	ipv4AddrClassA
	ipv4AddrClassAPrivate
	ipv4AddrClassLoopback
	ipv4AddrClassMax
)

type ipv4AddrGenerator func() (string, error)

var ipv4AddrGenerators = []ipv4AddrGenerator{
	ipv4AddrClassAny:      getAnyIpv4,
	ipv4AddrClassA:        getClassAIpv4,
	ipv4AddrClassAPrivate: getClassAPrivateIpv4,
	ipv4AddrClassLoopback: getLoopbackIpv4,
}

type ipv4Gen struct {
	block       []string
	count       int
	class       ipv4AddrClass
	initialized bool
}

func newIpv4Generator() *ipv4Gen {
	return &ipv4Gen{}
}

func (ipv4Gen *ipv4Gen) initIpv4Block(blockCount int, addrClass ipv4AddrClass) (err error) {
	if addrClass <= ipv4AddrClassMin || addrClass >= ipv4AddrClassMax {
		return fmt.Errorf("invalid address type %v", addrClass)
	}

	if ipv4Gen.initialized {
		return nil
	}

	ipv4Gen.class = addrClass
	ipv4Gen.count = blockCount
	ipv4Gen.block = make([]string, ipv4Gen.count)

	ipv4AddrGeneratorHandler := ipv4AddrGenerators[addrClass]
	for i := 0; i < blockCount; i++ {
		ipv4Addr, err := ipv4AddrGeneratorHandler()
		if err != nil {
			return err
		}

		ipv4Gen.block[i] = ipv4Addr
	}

	ipv4Gen.initialized = true
	return nil
}

func getAnyIpv4() (string, error) {
	octets := make([]int, 4)

	octets[0], _ = genIpv4OctetWithExcludeList(
		ipv4MinOctet,
		ipv4MaxOctet,
		[]int{ipv4MinOctet},
	)

	for oi := 1; oi < 4; oi++ {
		octets[oi], _ = genIpv4Octet(ipv4MinOctet, ipv4MaxOctet)
	}

	return getIpv4StringFromOctets(octets), nil
}

func getClassAIpv4() (string, error) {
	octets := make([]int, 4)

	octets[0], _ = genIpv4OctetWithExcludeList(
		ipv4MinOctet,
		ipv4ClassAMaxOctet,
		[]int{ipv4MinOctet, ipv4ClassAPrivateFirstOctet},
	)

	for oi := 1; oi < 4; oi++ {
		octets[oi], _ = genIpv4Octet(ipv4MinOctet, ipv4MaxOctet)
	}

	return getIpv4StringFromOctets(octets), nil
}

func getClassAPrivateIpv4() (string, error) {
	octets := make([]int, 4)

	octets[0] = ipv4ClassAPrivateFirstOctet

	for oi := 1; oi < 4; oi++ {
		octets[oi], _ = genIpv4Octet(ipv4MinOctet, ipv4MaxOctet)
	}

	return getIpv4StringFromOctets(octets), nil
}

func getLoopbackIpv4() (string, error) {
	octets := make([]int, 4)

	octets[0] = ipv4LoopbackFirstOctet

	for oi := 1; oi < 4; oi++ {
		octets[oi], _ = genIpv4Octet(ipv4MinOctet, ipv4MaxOctet)
	}

	return getIpv4StringFromOctets(octets), nil
}

func genIpv4Octet(min, max int) (int, error) {
	return genIpv4OctetWithExcludeList(min, max, []int{})
}

func genIpv4OctetWithExcludeList(min, max int, excludeList []int) (int, error) {
	if max < 0 {
		return 0, fmt.Errorf("invalid max: cannot be negative")
	}

	excludeMap := make(map[int]bool)
	for _, exclude := range excludeList {
		excludeMap[exclude] = true
	}

	octet := rand.Intn(max + 1)
	if octet < min {
		octet += rand.Intn(max - min)
	}

	if _, exists := excludeMap[octet]; exists {
		for octet = min; octet <= max; octet++ {
			if _, exists := excludeMap[octet]; !exists {
				break
			}
		}
	}

	return octet, nil
}

func getIpv4StringFromOctets(octets []int) (ipv4Addr string) {
	return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
}

type ipv4Validator func(string) error

var ipv4Validators = []ipv4Validator{
	ipv4AddrClassAny:      validateClassAny,
	ipv4AddrClassA:        validateClassA,
	ipv4AddrClassAPrivate: validateClassAPrivate,
	ipv4AddrClassLoopback: validateClassLoopback,
}

func (ipv4Gen *ipv4Gen) validateIpv4Address(ipv4Addr string) (err error) {
	if ipv4Gen.class <= ipv4AddrClassMin || ipv4Gen.class >= ipv4AddrClassMax {
		return fmt.Errorf("invalid address class %v", ipv4Gen.class)
	}

	return ipv4Validators[ipv4Gen.class](ipv4Addr)
}

func validateClassAny(ipv4Addr string) (err error) {
	nip := net.ParseIP(ipv4Addr).To4()
	if nil == nip {
		return fmt.Errorf("invalid ip address %v", ipv4Addr)
	}

	if nip[0] <= ipv4MinOctet {
		return fmt.Errorf("invalid ip address %v starting with octet 0", ipv4Addr)
	}

	return nil
}

func validateClassA(ipv4Addr string) (err error) {
	nip := net.ParseIP(ipv4Addr).To4()
	if nil == nip {
		return fmt.Errorf("invalid ip address %v", ipv4Addr)
	}

	if nip[0] <= ipv4MinOctet || nip[0] > ipv4ClassAMaxOctet {
		return fmt.Errorf("not a class A ip address %v", ipv4Addr)
	}

	if nip[0] == ipv4ClassAPrivateFirstOctet {
		return fmt.Errorf("class A private ip address %v", ipv4Addr)
	}

	return nil
}

func validateClassAPrivate(ipv4Addr string) (err error) {
	nip := net.ParseIP(ipv4Addr).To4()
	if nil == nip {
		return fmt.Errorf("invalid ip address %v", ipv4Addr)
	}

	if nip[0] != ipv4ClassAPrivateFirstOctet {
		return fmt.Errorf("not a class A private ip address %v", ipv4Addr)
	}

	return nil
}

func validateClassLoopback(ipv4Addr string) (err error) {
	nip := net.ParseIP(ipv4Addr).To4()
	if nil == nip {
		return fmt.Errorf("invalid ip address %v", ipv4Addr)
	}

	if nip[0] != ipv4LoopbackFirstOctet {
		return fmt.Errorf("not a loopback ip address %v", ipv4Addr)
	}

	return nil
}
