package ts3

import (
	"errors"
	"time"
)

// Pool manages a pool of TS3 clients for a single server.
type Pool struct {
	factory func() (*Client, error)
	clients chan *Client   // Idle clients
	sem     chan struct{}  // Semaphore for total clients
	max     int
}

// NewPool creates a new connection pool.
func NewPool(factory func() (*Client, error), max int) *Pool {
	return &Pool{
		factory: factory,
		clients: make(chan *Client, max),
		sem:     make(chan struct{}, max),
		max:     max,
	}
}

// Get retrieves a client from the pool or creates a new one.
func (p *Pool) Get() (*Client, error) {
	// 1. Try to get idle client
	select {
	case client := <-p.clients:
		return client, nil
	default:
		// No idle clients, proceed to create or wait
	}

	// 2. Try to create new if capacity allows
	select {
	case p.sem <- struct{}{}:
		// Slot acquired
		c, err := p.factory()
		if err != nil {
			<-p.sem // Release slot on failure
			return nil, err
		}
		return c, nil
	default:
		// Capacity full, wait for idle client or free slot
		select {
		case client := <-p.clients:
			return client, nil
		case p.sem <- struct{}{}:
			// Slot became free (e.g. from Discard)
			c, err := p.factory()
			if err != nil {
				<-p.sem
				return nil, err
			}
			return c, nil
		case <-time.After(10 * time.Second):
			return nil, errors.New("timeout waiting for connection from pool")
		}
	}
}

// Put returns a client to the pool.
func (p *Pool) Put(c *Client) {
	select {
	case p.clients <- c:
		// returned to pool (slot in sem is still held)
	default:
		// pool full? Should not happen if max logic is correct.
		// If it happens, discard.
		c.Close()
		select {
		case <-p.sem:
		default:
		}
	}
}

// Discard closes a client and releases its slot.
func (p *Pool) Discard(c *Client) {
	c.Close()
	select {
	case <-p.sem:
		// Slot released
	default:
		// Should not happen
	}
}

// Close closes all clients in the pool.
func (p *Pool) Close() {
	close(p.clients)
	for c := range p.clients {
		c.Close()
	}
	// Note: p.sem is not closed as concurrent Get might panic sending to closed channel.
	// We assume Close is called when system shuts down.
}
