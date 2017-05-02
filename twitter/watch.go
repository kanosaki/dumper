package twitter

import (
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/kanosaki/go-twitter/twitter"
)

type Watcher interface {
	Refresh(ctx context.Context) error
	Key() string
}

type ListWatch struct {
	client   *twitter.Client
	target   TargetList
	tw       *Twitter
	latestID int64
}

func (lw *ListWatch) Key() string {
	return fmt.Sprintf("/twitter/list/%s/%s", lw.target.OwnerScreenName, lw.target.Slug)
}

func flipTweets(s []twitter.Tweet) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (lw *ListWatch) Refresh(ctx context.Context) error {
	log.Debugf("Twitter: %v", lw.Key())
	tweets, resp, err := lw.client.Lists.Statuses(&twitter.ListStatusesParams{
		Slug:            lw.target.Slug,
		OwnerScreenName: lw.target.OwnerScreenName,
		IncludeEntities: true,
		SinceID:         lw.latestID,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// flip tweets to order tweets be first
	flipTweets(tweets)
	for _, tw := range tweets {
		if lw.latestID < tw.ID {
			lw.latestID = tw.ID
		}
	}
	lw.tw.Next(lw.Key(), tweets)
	return nil
}
