package timeline

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timeline item
type Item struct {
	ID        int64 `json:"id"` // Sequence number in Timeline
	Caption   string `json:"caption"`
	Thumbnail string `json:"thumbnail"`
	Timestamp time.Time `json:"timestamp"`
	TopicID   int `json:"timelineID"`
	TopicKey  string `json:"-"`
	OriginKey int64 `json:"key"` // ID for each timeline
	Meta      map[string]interface{} `json:"meta"`
}

func (i *Item) EncodeMeta() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Item) String() string {
	return fmt.Sprintf("Item{%d, %v, %d, %d}", i.ID, i.Caption, i.TopicID, i.OriginKey)
}

type Publishing struct {
	Topic string
	Item  *Item
}
