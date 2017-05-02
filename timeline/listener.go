package timeline

import "sync"

type Listener struct {
	Key        string
	C          chan *Item
	closed     bool
	overflowed uint32
	fetchMu    sync.Mutex
	pushMu     sync.Mutex
}

func (l *Listener) Push(i *Item) {
	l.pushMu.Lock()
	defer l.pushMu.Unlock()
	if len(l.C) == cap(l.C) {
		l.overflowed = (l.overflowed + 1) % 0xFFFFFFFF
		return
	}
	l.C <- i
}

func (l *Listener) Fetch(limit int) []*Item {
	l.fetchMu.Lock()
	defer l.fetchMu.Unlock()
	chLen := len(l.C)
	// MIN(limit, chLen)
	if chLen < limit {
		limit = chLen
	}
	// but limit = 0, infinite
	if limit == 0 && chLen != 0 {
		limit = chLen
	}
	ret := make([]*Item, 0, limit)
	for i := 0; i < limit; i++ {
		ret = append(ret, <-l.C)
	}
	return ret
}

func (l *Listener) Close() {
	l.closed = true
}
