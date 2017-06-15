package timeline

import (
	"container/list"
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	DefaultListenerBuffer = 200
	ListenerCleanupThresh = 4
	DefaultHistorySize    = 100
	ErrNoTopic            = errors.New("No matched topic")
)

type Service struct {
	ListenerBuffer int
	HistorySize    int
	topics         map[string]*Topic
	topicKeys      []string // list of topics, sorted
	topicsMu       sync.RWMutex
	persistent     Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		ListenerBuffer: DefaultListenerBuffer,
		HistorySize:    DefaultHistorySize,
		topics:         make(map[string]*Topic),
		persistent:     storage,
	}
}

func (s *Service) NewTopic(origin, key string) error {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	if _, ok := s.topics[key]; ok {
		return nil
	}
	var topicID int
	if s.persistent != nil {
		ctx := context.Background()
		originID, err := s.persistent.OriginID(ctx, origin, true)
		if err != nil {
			return err
		}
		topicID, err = s.persistent.TopicID(ctx, key, originID, true)
		if err != nil {
			return err
		}
	}
	s.topicKeys = append(s.topicKeys, key)
	sort.Strings(s.topicKeys)
	t := &Topic{
		Key:     key,
		s:       s,
		history: list.New(),
		ID:      topicID,
	}
	s.topics[key] = t
	// TODO: Consider updating listeners topic
	return nil
}

func (s *Service) Publish(topic string, item ... *Item) error {
	s.topicsMu.RLock()
	defer s.topicsMu.RUnlock()
	ctx := context.Background()
	for _, it := range item {
		t, ok := s.topics[topic]
		if !ok {
			return ErrNoTopic
		}
		it.TopicID = t.ID
	}
	if s.persistent != nil {
		if _, err := s.persistent.Insert(ctx, item...); err != nil {
			return err
		}
	}
	for _, it := range item {
		published := make(map[*Listener]struct{})
		for i := 0; i < len(s.topicKeys); i++ {
			// reverse loop --> seek longest match topic
			k := s.topicKeys[len(s.topicKeys)-i-1]
			if topic == k {
				s.topics[k].publish(it, published)
			}
		}
	}
	return nil
}

func (s *Service) Topics(prefix string) (ret []*Topic) {
	s.topicsMu.RLock()
	defer s.topicsMu.RUnlock()
	for _, k := range s.topicKeys {
		if strings.HasPrefix(k, prefix) {
			ret = append(ret, s.topics[k])
		}
	}
	return
}

func (s *Service) Fetch(ctx context.Context, q *Query) ([]*Item, error) {
	return s.persistent.Select(ctx, q)
}

func (s *Service) Listen(key string) (*Listener, error) {
	lis := &Listener{
		Key: key,
		C:   make(chan *Item, s.ListenerBuffer),
	}
	s.topicsMu.RLock()
	found := false
	defer s.topicsMu.RUnlock()
	for i := 0; i < len(s.topicKeys); i++ {
		// reverse loop --> seek longest match topic
		k := s.topicKeys[len(s.topicKeys)-i-1]
		if strings.HasPrefix(k, key) {
			s.topics[k].addListener(lis)
			found = true
		}
	}
	if !found {
		return nil, ErrNoTopic
	}
	return lis, nil
}

func (s *Service) PrintStatus() {
	for i := 0; i < len(s.topicKeys); i++ {
		// reverse loop --> seek longest match topic
		k := s.topicKeys[len(s.topicKeys)-i-1]
		s.topics[k].PrintStatus()
	}
}
