package prefix_tree

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

const (
	v4GenCount = 256
)

func validatev4Addr(t *testing.T, saddr string, exip string, exmask string, negate bool) {
	nip, nmask, err := testgetv4Addr(saddr)
	if (nil == err) == negate {
		t.Fatalf("getv4Addr: error validating %s and negate %v", saddr, negate)
	}

	if !negate && (nip.String() != exip || exmask != nmask.String()) {
		t.Fatalf("getv4Addr: invalid values when validating %s and negate %v", saddr, negate)
	}
}

func basicV4Tests(t *testing.T) {
	validatev4Addr(t, "", "", "", true)
	validatev4Addr(t, "192.168.128.40/33", "", "", true)
	validatev4Addr(t, "256.168.128.40/32", "", "", true)
	validatev4Addr(t, "2001:db8:a0b:12f0::1", "", "", true)
	validatev4Addr(t, "2001:db8:a0b:12f0::1/64", "", "", true)
}

func randomV4Tests(t *testing.T) {
	v4t := NewV4Tree()

	for class := ipv4AddrClassMin + 1; class < ipv4AddrClassMax; class++ {
		v4gen := newIpv4Generator()

		err := v4gen.initIpv4Block(v4GenCount, class)
		if err != nil {
			t.Fatalf("failed to initialize v4 ip addresses for class %v", class)
		}

		for i := 0; i < v4gen.count; i++ {
			addr := v4gen.block[i]

			validatev4Addr(t, addr, addr, net.CIDRMask(32, 32).String(), false)

			for m := 32; m > 0; m-- {
				cidr := addr + "/" + fmt.Sprint(m)

				_, netaddr, err := net.ParseCIDR(cidr)
				if err != nil {
					t.Fatalf("failed to parse %s, %v", cidr, err)
				}

				validatev4Addr(t, cidr, netaddr.IP.String(), netaddr.Mask.String(), false)
			}
		}

		for i := 0; i < v4gen.count; i++ {
			res, err := v4t.Insert(v4gen.block[i], nil)
			if err != nil || res != Ok {
				t.Fatalf("Failed to insert %s", v4gen.block[i])
			}

			res, _, err = v4t.Search(v4gen.block[i])
			if err != nil || res != Match {
				t.Fatalf("Failed to find %s", v4gen.block[i])
			}

			res, _, err = v4t.SearchExact(v4gen.block[i] + "/32")
			if err != nil || res != Match {
				t.Fatalf("Failed to find (exact) %s", v4gen.block[i]+"/32")
			}

			res, err = v4t.Insert(v4gen.block[i]+"/32", nil)
			if err != nil || res != Dup {
				t.Fatalf("Failed to recognize %s as duplicate", v4gen.block[i]+"/32")
			}

			res, err = v4t.Insert(v4gen.block[i], nil)
			if err != nil || res != Dup {
				t.Fatalf("Failed to recognize %s as duplicate", v4gen.block[i])
			}

			res, _, err = v4t.Delete(v4gen.block[i] + "/32")
			if err != nil || res != Match {
				t.Fatalf("Failed to delete %s", v4gen.block[i]+"/32")
			}

			res, _, err = v4t.Search(v4gen.block[i] + "/32")
			if nil == err || res != Err {
				t.Fatalf("Found non existent key %s", v4gen.block[i]+"/32")
			}

			res, _, err = v4t.SearchExact(v4gen.block[i])
			if nil == err || res != Err {
				t.Fatalf("Found (exact) non existent key %s", v4gen.block[i])
			}

			res, _, err = v4t.Delete(v4gen.block[i])
			if nil == err || res != Err {
				t.Fatalf("Deleted non existent key %s", v4gen.block[i])
			}
		}
	}
}

func extendedV4Tests(t *testing.T) {
	v4t := NewV4Tree()

	elemsMap := make(map[string]bool)

	prefix := "192.168"

	for i := 0; i < 255; i++ {
		for j := 0; j < 255; j++ {
			for m := 32; m > 24; m-- {
				cidr := fmt.Sprintf("%s.%d.%d/%d", prefix, i, j, m)

				ip, mask, _ := getv4Addr(cidr)

				value := fmt.Sprintf("%s/%s", ip.String(), mask.String())

				if _, exists := elemsMap[value]; !exists {
					res, err := v4t.Insert(cidr, value)
					if err != nil || res != Ok {
						t.Fatalf("Failed to insert %s, result = %v, %v", cidr, res, err)
					}

					elemsMap[value] = true
				} else {
					res, err := v4t.Insert(cidr, value)
					if err != nil || res != Dup {
						t.Fatalf("Failed to identify duplicate %s, result = %v, %v", cidr, res, err)
					}
				}
			}
		}
	}

	for i := 0; i < 255; i++ {
		for j := 0; j < 255; j++ {
			for m := 32; m > 24; m-- {
				cidr := fmt.Sprintf("%s.%d.%d/%d", prefix, i, j, m)

				ip, mask, _ := getv4Addr(cidr)

				value := fmt.Sprintf("%s/%s", ip.String(), mask.String())

				if _, exists := elemsMap[value]; !exists {
					continue
				}

				res, saved, err := v4t.SearchExact(cidr)
				if nil == err && Match == res {
					if saved.(string) != value {
						t.Fatalf("Search failed for %s, returned %s, expected %s", cidr, saved.(string), value)
					}
				} else {
					t.Fatalf("Search failed for %s, %v/%v", cidr, res, err)
				}

				res, saved, err = v4t.Delete(cidr)
				if nil == err && Match == res {
					if saved.(string) != value {
						t.Fatalf("Delete failed for %s, returned %s, expected %s", cidr, saved.(string), value)
					}
				} else {
					t.Fatalf("Delete failed for %s, %v/%v", cidr, res, err)
				}

				delete(elemsMap, value)
			}
		}
	}
}

func TestV4(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	basicV4Tests(t)
	randomV4Tests(t)
	extendedV4Tests(t)
}
