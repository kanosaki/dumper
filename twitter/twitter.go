package twitter

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/oauth1"
	"github.com/kanosaki/dumper/core"
	"github.com/kanosaki/dumper/pkg/balancer"
	"github.com/kanosaki/dumper/timeline"
	"github.com/kanosaki/go-twitter/twitter"
)

var (
	OriginTwitterTweet = "twitter.tweet"
)

type Twitter struct {
	conf   TwitterConf
	client *twitter.Client
	c      *core.Context
	wbl    *balancer.RoundRobbin
}

func (t *Twitter) Init(id string, c *core.Context) error {
	t.c = c
	if err := c.Config().Unmarshal("modules/twitter", &t.conf); err != nil {
		return err
	}
	conf := t.conf
	config := oauth1.NewConfig(conf.ConsumerKey, conf.ConsumerSecret)
	token := oauth1.NewToken(conf.Token, conf.TokenSecret)
	t.client = twitter.NewClient(config.Client(context.Background(), token))
	t.wbl = balancer.NewRoundRobbin(nil)
	for _, wl := range conf.WatchLists {
		lw := &ListWatch{
			client: t.client,
			target: wl,
			tw:     t,
		}
		c.Timeline().NewTopic(OriginTwitterTweet, lw.Key())
		t.wbl.Add(lw.Refresh)
	}
	return nil
}

func (t *Twitter) Start(c *core.Context) error {
	t.c = c
	tick := time.NewTicker(5 * time.Minute)
	log.Debugf("Twitter START")
	for {
		select {
		case <-tick.C:
			log.Debugf("Twitter FIRE")
			ctx := context.Background()
			if err := t.wbl.Fire(ctx); err != nil {
				log.Errorf("Twitter: %v", err)
			}
		}
	}
	return nil
}

func (t *Twitter) Next(key string, tweet []twitter.Tweet) {
	tl := t.c.Timeline()
	for _, tw := range tweet {
		if len(tw.Entities.Media) == 0 {
			continue
		}
		media := tw.Entities.Media
		thumbnailURL := fmt.Sprintf("%s:thumb", media[0].MediaURL)
		err := tl.Publish(key, &timeline.Item{
			Caption:   fmt.Sprintf("%s / %s", tw.User.ScreenName, tw.Text),
			OriginKey: tw.ID,
			Thumbnail: thumbnailURL,
		})
		if err != nil {
			log.Errorf("Failed to publish twitter item: %v", err)
		}
	}
}

func (t *Twitter) Close() error {
	return nil
}

func (t *Twitter) Status() core.ModuleStatus {
	return core.Status.OK()
}

type TwitterConf struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	Token          string `yaml:"token"`
	TokenSecret    string `yaml:"token_secret"`
	WatchLists     []TargetList `yaml:"watch_lists"`
}

type TargetList struct {
	OwnerScreenName string `yaml:"owner"`
	Slug            string `yaml:"slug"`
}
