package prefix_tree

import (
	"context"
	"net"
	"testing"
)

func validatev6Addr(t *testing.T, saddr string, exip string, exmask string, negate bool) {
	nip, nmask, err := testgetv6Addr(saddr)
	if (nil == err) == negate {
		t.Fatalf("getv6Addr: error validating %s and negate %v", saddr, negate)
	}

	if !negate && (nip.String() != exip || exmask != nmask.String()) {
		t.Fatalf("getv6Addr: invalid values when validating %s and negate %v", saddr, negate)
	}
}

func TestV6(t *testing.T) {
	validatev6Addr(t, "", "", "", true)
	validatev6Addr(t, "2001:db8:a0b:12f0::1", "2001:db8:a0b:12f0::1", net.CIDRMask(128, 128).String(), false)
	validatev6Addr(t, "2001:db8:a0b:12f0::1/128", "2001:db8:a0b:12f0::1", net.CIDRMask(128, 128).String(), false)
	validatev6Addr(t, "2001:db8:a0b:12f0::1/64", "2001:db8:a0b:12f0::", net.CIDRMask(64, 128).String(), false)
	validatev6Addr(t, "2001:db8:a0b:12f0::1/48", "2001:db8:a0b::", net.CIDRMask(48, 128).String(), false)
	validatev6Addr(t, "2001:db8:a0b:12f0::1/42", "2001:db8:a00::", net.CIDRMask(42, 128).String(), false)
	validatev6Addr(t, "59fb::1005:cc57:6571/128", "59fb::1005:cc57:6571", net.CIDRMask(128, 128).String(), false)
	validatev6Addr(t, "56fe::2159:5bbc::6594", "", "", true)
	validatev6Addr(t, "192.168.128.40", "", "", true)
	validatev6Addr(t, "192.168.128.40/32", "", "", true)
}

// BenchmarkV6TreeInsert benchmarks V6Tree.Insert
func BenchmarkV6TreeInsert(b *testing.B) {
	ctx := context.Background()
	v6tree := NewV6Tree[int]()
	addresses := generateIPv6Addresses(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ival := i
		v6tree.Insert(ctx, addresses[i], &ival)
	}
}

// BenchmarkV6TreeSearch benchmarks V6Tree.Search
func BenchmarkV6TreeSearch(b *testing.B) {
	ctx := context.Background()
	v6tree := NewV6Tree[int]()
	addresses := generateIPv6Addresses(b.N)

	// Pre-populate
	for i := 0; i < b.N; i++ {
		ival := i
		v6tree.Insert(ctx, addresses[i], &ival)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v6tree.Search(ctx, addresses[i])
	}
}
