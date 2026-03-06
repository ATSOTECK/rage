package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/ATSOTECK/rage/pkg/rage"
)

// Handle is returned by Borrow and must be passed to Return.
type Handle struct {
	State *rage.State
}

// StatePool is a channel-based pool of pre-initialized RAGE states.
type StatePool struct {
	entries  chan *Handle
	name     string
	createMu *sync.Mutex
}

// New creates a pool of RAGE states. createMu must be shared across all pools
// to serialize state creation (rage.NewStateWithModules touches global state).
func New(name string, size int, code *rage.Code, createMu *sync.Mutex) (*StatePool, error) {
	p := &StatePool{
		entries:  make(chan *Handle, size),
		name:     name,
		createMu: createMu,
	}

	for i := range size {
		createMu.Lock()
		state := rage.NewStateWithModules(rage.WithModule(rage.ModuleMath), rage.WithModule(rage.ModuleString))
		if _, err := state.Execute(code); err != nil {
			createMu.Unlock()
			close(p.entries)
			for h := range p.entries {
				h.State.Close()
			}
			return nil, fmt.Errorf("pool %s: create state %d: %w", name, i, err)
		}
		createMu.Unlock()
		p.entries <- &Handle{State: state}
	}

	return p, nil
}

// Borrow blocks until a state is available or ctx is cancelled.
func (p *StatePool) Borrow(ctx context.Context) (*Handle, error) {
	select {
	case h := <-p.entries:
		return h, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Return puts a state back in the pool.
func (p *StatePool) Return(h *Handle) {
	p.entries <- h
}

// Close drains the pool and closes all states.
func (p *StatePool) Close() {
	close(p.entries)
	for h := range p.entries {
		h.State.Close()
	}
}
