package timeline

import (
	"container/list"
	"fmt"
	"sync"
)

type Topic struct {
	Key         string
	ID          int
	s           *Service
	history     *list.List
	historyMu   sync.Mutex
	listeners   []*Listener
	listenersMu sync.Mutex
}

func (t *Topic) pushHistory(item *Item) {
	t.historyMu.Lock()
	defer t.historyMu.Unlock()
	t.history.PushBack(item)
	for t.history.Len() >= t.s.HistorySize {
		t.history.Remove(t.history.Front())
	}
}

func (t *Topic) publish(item *Item, published map[*Listener]struct{}) {
	closedCount := 0
	for _, l := range t.listeners {
		if l.closed {
			closedCount++
			continue
		}
		if _, ok := published[l]; ok {
			continue
		}
		l.Push(item)
		published[l] = struct{}{}
	}
	t.pushHistory(item)
	if closedCount > ListenerCleanupThresh {
		t.listenersMu.Lock()
		defer t.listenersMu.Unlock()
		cleanedListeners := make([]*Listener, 0, len(t.listeners)-closedCount)
		for _, l := range t.listeners {
			if l.closed {
				continue
			}
			cleanedListeners = append(cleanedListeners, l)
		}
		t.listeners = cleanedListeners
	}
}

func (t *Topic) addListener(l *Listener) {
	t.listenersMu.Lock()
	defer t.listenersMu.Unlock()
	t.listeners = append(t.listeners, l)
}

func (t *Topic) PrintStatus() {
	fmt.Printf("%v history: %d, listeners: %d\n", t.Key, t.history.Len(), len(t.listeners))
	for _, l := range t.listeners {
		fmt.Printf("\t%v\n", l.Key)
	}
}
