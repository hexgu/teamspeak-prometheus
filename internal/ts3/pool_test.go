package ts3

import (
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	factory := func() (*Client, error) {
		return &Client{}, nil
	}

	p := NewPool(factory, 2)

	// Get 1
	c1, err := p.Get()
	if err != nil {
		t.Fatalf("Get 1 failed: %v", err)
	}

	// Get 2
	c2, err := p.Get()
	if err != nil {
		t.Fatalf("Get 2 failed: %v", err)
	}

	// Get 3 (should block, simulate via goroutine)
	start := time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		p.Put(c1)
	}()

	c3, err := p.Get()
	if err != nil {
		t.Fatalf("Get 3 failed: %v", err)
	}
	if time.Since(start) < 100*time.Millisecond {
		t.Errorf("Get 3 should have blocked")
	}

	// c3 should be c1 recycled
	if c3 != c1 {
		t.Errorf("Expected c3 to be c1, got %p vs %p", c3, c1)
	}

	// Discard c2
	p.Discard(c2)

	// Get 4 (should create new because slot freed)
	c4, err := p.Get()
	if err != nil {
		t.Fatalf("Get 4 failed: %v", err)
	}
	// c4 should be new
	if c4 == c1 || c4 == c2 {
		t.Errorf("Expected c4 to be new, but got existing")
	}
}
