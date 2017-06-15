package timeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
)

func printTimelines(db Storage) error {
	// Debug print contents of database
	rows, err := db.DB().Query(`SELECT timeline.id, topic_id, caption, thumbnail, origin_key, topic.key FROM timeline JOIN topic ON topic.id = topic_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var caption, thumbnail, topicName string
		var originKey, topicID, id int64
		if err := rows.Scan(&id, &topicID, &caption, &thumbnail, &originKey, &topicName); err != nil {
			return err
		}
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", id, topicID, caption, thumbnail, originKey, topicName)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return err
	}
	return nil
}

func mapOriginKey(items []*Item) []int64 {
	ret := make([]int64, 0, len(items))
	for _, item := range items {
		ret = append(ret, item.OriginKey)
	}
	return ret
}

func TestSQLiteStorage(t *testing.T) {
	a := assert.New(t)
	db, err := NewStorage("memory", "")
	if err != nil {
		t.Error(err)
		return
	}
	now := time.Now()
	ctx := context.Background()
	origin1, err := db.OriginID(ctx, "twitter/timeline/status", true)
	if err != nil {
		t.Error(err)
		return
	}
	topicNames := []string{"/user1/List1", "/user1/List2"}
	topics := make(map[string]int)
	for _, topicName := range topicNames {
		topicID, err := db.TopicID(ctx, topicName, origin1, true)
		if err != nil {
			t.Error(err)
			return
		}
		topics[topicName] = topicID
	}
	items := []*Item{
		{TopicID: topics[topicNames[0]], OriginKey: 11, Thumbnail: "http://foo/1a.jpg", Caption: "Tweet 1A", Timestamp: now.Add(1 * time.Second)},
		{TopicID: topics[topicNames[0]], OriginKey: 12, Thumbnail: "http://foo/1b.jpg", Caption: "Tweet 1B", Timestamp: now.Add(2 * time.Second)},
		{TopicID: topics[topicNames[0]], OriginKey: 13, Thumbnail: "http://foo/1c.jpg", Caption: "Tweet 1C", Timestamp: now.Add(3 * time.Second)},
		{TopicID: topics[topicNames[0]], OriginKey: 14, Thumbnail: "http://foo/1d.jpg", Caption: "Tweet 1D", Timestamp: now.Add(4 * time.Second)},
		{TopicID: topics[topicNames[1]], OriginKey: 21, Thumbnail: "http://foo/2a.jpg", Caption: "Tweet 2A", Timestamp: now.Add(1 * time.Second)},
		{TopicID: topics[topicNames[1]], OriginKey: 22, Thumbnail: "http://foo/2b.jpg", Caption: "Tweet 2B", Timestamp: now.Add(2 * time.Second)},
		{TopicID: topics[topicNames[1]], OriginKey: 23, Thumbnail: "http://foo/2c.jpg", Caption: "Tweet 2C", Timestamp: now.Add(3 * time.Second)},
		{TopicID: topics[topicNames[1]], OriginKey: 24, Thumbnail: "http://foo/2d.jpg", Caption: "Tweet 2D", Timestamp: now.Add(4 * time.Second)},
	}
	for _, item := range items {
		_, err = db.Insert(ctx, item)
		if err != nil {
			t.Error(err)
			return
		}
	}

	// print db content
	//if err := printTimelines(db); err != nil {
	//	t.Error(err)
	//	return
	//}

	testPairs := []struct {
		originKeys []int64;
		query      *Query
	}{
		// Empty
		{[]int64{24, 23, 22, 21, 14, 13, 12, 11}, &Query{}},

		// Simple
		{[]int64{24, 23}, &Query{Limit: 2}},
		{[]int64{24, 23, 22, 21}, &Query{Topics: []string{"/user1/List2"}}},
		{[]int64{24, 23, 22, 21, 14, 13, 12, 11}, &Query{Topics: []string{"/user1/List1", "/user1/List2"}}},
		{[]int64{22, 21, 14, 13, 12, 11}, &Query{MaxID: 6}},
		{[]int64{24, 14, 23, 13}, &Query{After: now.Add(3 * time.Second)}},
		{[]int64{24, 23, 22}, &Query{MinID: 6}},
		{[]int64{13, 23, 12, 22, 11, 21}, &Query{Before: now.Add(3 * time.Second)}},

		// Compound
		{[]int64{22, 21, 14, 13}, &Query{MaxID: 6, Limit: 4}},
		{[]int64{23, 22}, &Query{MinID: 6, Limit: 2}},
		{[]int64{22, 21, 14}, &Query{MaxID: 6, MinID: 4}},
		{[]int64{22, 21}, &Query{MaxID: 6, MinID: 4, Limit: 2}},
		{[]int64{13, 23, 12, 22}, &Query{Before: now.Add(3 * time.Second), Limit: 4}},
		{[]int64{23, 22}, &Query{Before: now.Add(3 * time.Second), Limit: 2, Topics: []string{"/user1/List2"}}},
	}
	for _, pair := range testPairs {
		ps, err := db.Select(ctx, pair.query)
		if err != nil {
			t.Error(err, pp.Sprint(pair.query))
			return
		}
		a.Equal(pair.originKeys, mapOriginKey(ps), pp.Sprint(pair.query))
	}
}
