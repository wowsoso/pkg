package limiter

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	res := newStore(0)

	if res.size != MINSIZE {
		t.Fatalf("store size should return %v, got: %v", MINSIZE, res.size)
	}
	if res.clip != res.size>>8 {
		t.Fatalf("store clip should return %v, got: %v", res.size>>8, res.clip)
	}

	res = newStore(MINSIZE + 1)

	if res.size != MINSIZE+1 {
		t.Fatalf("store size should return %v, got: %v", MINSIZE+1, res.size)
	}
	if res.clip != res.size>>8 {
		t.Fatalf("store size should return %v, got: %v", res.size>>8, res.clip)
	}

	res = newStore(300000)

	if res.size != 300000 {
		t.Fatalf("store size should return %v, got: %v", 300000, res.size)
	}
	if res.clip != MAXCLIP {
		t.Fatalf("store size should return %v, got: %v", MAXCLIP, res.clip)
	}

}

func TestStoreMethodGetSet(t *testing.T) {
	s := newStore(1024)
	s.size = 2
	s.clip = 2

	s.getSet("test1", 1, 1)
	s.getSet("test2", 1, 1)
	s.getSet("test3", 1, 1)

	if s.l.Len() != 1 && len(s.tokenBuckets) != 1 {
		t.Fatalf("length should be 1, got: %v, %v", s.l.Len(), len(s.tokenBuckets))
	}

	key := s.l.Back().Value.(*node).key
	if key != "test3" {
		t.Fatalf("key should be test1, got: %v, %v", key)
	}

	s.getSet("test3", 1, 1)

	if s.l.Len() != 1 && len(s.tokenBuckets) != 1 {
		t.Fatalf("length should be 1, got: %v, %v", s.l.Len(), len(s.tokenBuckets))
	}

	key = s.l.Back().Value.(*node).key
	if key != "test3" {
		t.Fatalf("key should be test1, got: %v, %v", key)
	}

}

func TestStoreMethodLimiter(t *testing.T) {
	config := Config{
		Max:  1000,
		TTL:  1 * time.Second,
		Size: MINSIZE,
	}

	l := NewLimiter(config)

	l.LimitByRequests("0.0.0.0", "/")
	l.LimitByBytes("0.0.0.0", 2)
}
