package prefix_tree

import (
	"context"
	"testing"
)

func TestTree_Insert_Search_Delete(t *testing.T) {
	ctx := context.Background()
	tr := NewTree()

	// basic insert
	key := []byte{0xC0, 0xA8, 0x01, 0x01} // 192.168.1.1
	mask := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	res, err := tr.Insert(ctx, key, mask, "value1")
	if err != nil || res != Ok {
		t.Fatalf("Insert failed: res=%v err=%v", res, err)
	}

	if tr.NumNodes != 1 {
		t.Fatalf("expected NumNodes 1 got %d", tr.NumNodes)
	}

	// duplicate insert returns Dup
	res, err = tr.Insert(ctx, key, mask, "value1-dup")
	if err != nil || res != Dup {
		t.Fatalf("expected Dup on duplicate insert, got res=%v err=%v", res, err)
	}

	// exact search must find value
	res, v, err := tr.SearchExact(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchExact failed: res=%v err=%v", res, err)
	}
	if v != "value1" {
		t.Fatalf("unexpected value: %v", v)
	}

	// partial search for same key should also find the value
	res, v, err = tr.SearchPartial(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchPartial failed: res=%v err=%v", res, err)
	}
	if v != "value1" {
		t.Fatalf("unexpected value: %v", v)
	}

	// delete should return the stored value and decrement NumNodes
	res, val, err := tr.Delete(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("Delete failed: res=%v err=%v", res, err)
	}
	if val != "value1" {
		t.Fatalf("Delete returned wrong value: %v", val)
	}
	if tr.NumNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.NumNodes)
	}

	// delete non-existent should return error
	res, _, err = tr.Delete(ctx, key, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error deleting non-existent key")
	}
}

func TestTree_Partial_Prefixs(t *testing.T) {
	ctx := context.Background()
	tr := NewTree()

	// insert a /24 prefix: 10.0.1.0/24 -> mask first three octets
	key := []byte{10, 0, 1, 0}
	mask := []byte{0xFF, 0xFF, 0xFF, 0x00} // /24

	res, err := tr.Insert(ctx, key, mask, "net-10-0-1")
	if err != nil || res != Ok {
		t.Fatalf("Insert prefix failed: %v %v", res, err)
	}

	// search for an address inside prefix should match partial search
	addr := []byte{10, 0, 1, 42}
	res, val, err := tr.SearchPartial(ctx, addr, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchPartial for address inside prefix failed: %v %v", res, err)
	}
	if val != "net-10-0-1" {
		t.Fatalf("unexpected partial match value: %v", val)
	}

	// exact search for the specific network key should also match
	res, val, err = tr.SearchExact(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("SearchExact for network key failed: %v %v", res, err)
	}
	if val != "net-10-0-1" {
		t.Fatalf("unexpected exact match value: %v", val)
	}

	// delete should return the stored value and decrement NumNodes
	res, val, err = tr.Delete(ctx, key, mask)
	if err != nil || res != Match {
		t.Fatalf("Delete prefix failed: %v %v", res, err)
	}
	if val != "net-10-0-1" {
		t.Fatalf("Delete returned wrong value for prefix: %v", val)
	}
	if tr.NumNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.NumNodes)
	}

	// search after delete should fail
	res, _, err = tr.SearchExact(ctx, key, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error searching deleted prefix")
	}

	// partial search for address inside deleted prefix should also fail
	res, _, err = tr.SearchPartial(ctx, addr, mask)
	if err == nil || res != Error {
		t.Fatalf("expected error searching address inside deleted prefix")
	}

	// Insert again to ensure tree is still functional
	res, err = tr.Insert(ctx, key, mask, "net-10-0-1-again")
	if err != nil || res != Ok {
		t.Fatalf("Re-insert prefix failed: %v %v", res, err)
	}

	// Delete with partial key should work
	res, val, err = tr.Delete(ctx, addr, mask)
	if err != nil {
		t.Fatalf("Delete with partial key failed: %v %v", res, err)
	}
	if val != "net-10-0-1-again" {
		t.Fatalf("Delete with partial key returned wrong value: %v", val)
	}
	if tr.NumNodes != 0 {
		t.Fatalf("expected NumNodes 0 got %d", tr.NumNodes)
	}
}

func TestTree_LockingHandlers(t *testing.T) {
	ctx := context.Background()
	// Very small test to ensure lock handlers are called
	called := struct{ r, ru, w, u int }{}

	rlock := func(_ context.Context) { called.r++ }
	runlock := func(_ context.Context) { called.ru++ }
	wlock := func(_ context.Context) { called.w++ }
	unlock := func(_ context.Context) { called.u++ }

	tr := NewTreeWithLockHandlers(rlock, runlock, wlock, unlock)

	key := []byte{1, 2, 3, 4}
	mask := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	// Insert should call write lock/unlock
	_, _ = tr.Insert(ctx, key, mask, "x")
	if called.w == 0 || called.u == 0 {
		t.Fatalf("expected write lock/unlock called, got w=%d u=%d", called.w, called.u)
	}

	// Search should call read lock/unlock
	_, _, _ = tr.SearchExact(ctx, key, mask)
	if called.r == 0 || called.ru == 0 {
		t.Fatalf("expected read lock/unlock called, got r=%d ru=%d", called.r, called.ru)
	}

	// SearchPartial should call read lock/unlock
	_, _, _ = tr.SearchPartial(ctx, key, mask)
	if called.r < 2 || called.ru < 2 {
		t.Fatalf("expected read lock/unlock called twice, got r=%d ru=%d", called.r, called.ru)
	}

	// Delete should call write lock/unlock
	_, _, _ = tr.Delete(ctx, key, mask)
	if called.w < 2 || called.u < 2 {
		t.Fatalf("expected write lock/unlock called twice, got w=%d u=%d", called.w, called.u)
	}
}
