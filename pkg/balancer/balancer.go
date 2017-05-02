package balancer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rubyist/circuitbreaker"
)

var (
	DefaultTimeout = 10 * time.Second
	ErrBreakerOpen = errors.New("No available callee")
)

type RoundRobbin struct {
	Timeout   time.Duration
	callees   []*callee
	calleesMu sync.Mutex
	opts      *circuit.Options
	next      int
}

type callee struct {
	cb *circuit.Breaker
	fn func(ctx context.Context) error
}

func NewRoundRobbin(opts *circuit.Options) *RoundRobbin {
	return &RoundRobbin{
		opts: opts,
	}
}

func (r *RoundRobbin) Add(fn func(ctx context.Context) error) {
	r.calleesMu.Lock()
	defer r.calleesMu.Unlock()
	r.callees = append(r.callees, &callee{
		cb: circuit.NewBreakerWithOptions(r.opts),
		fn: fn,
	})
}

func (r *RoundRobbin) Fire(ctx context.Context) error {
	r.calleesMu.Lock()
	var selectedCallee *callee
	start := r.next
	for i := start; i < start+len(r.callees); i++ {
		c := r.callees[i%len(r.callees)]
		if c.cb.Ready() {
			selectedCallee = c
			break
		}
	}
	r.next = (r.next + 1) % len(r.callees)
	r.calleesMu.Unlock()
	if selectedCallee == nil {
		return ErrBreakerOpen
	}
	timeout := r.Timeout
	if timeout.Nanoseconds() == 0 {
		timeout = DefaultTimeout
	}
	return selectedCallee.cb.CallContext(ctx, func() error {
		return selectedCallee.fn(ctx)
	}, timeout)
}
