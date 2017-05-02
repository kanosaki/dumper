package pixiv

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/kanosaki/dumper/core"
	"github.com/kanosaki/dumper/timeline"
	"github.com/kanosaki/gopixiv"
	"golang.org/x/net/context"
)

const (
	DailyRankingKey = "/pixiv/ranking/daily"
	OriginPixivWork = "pixiv/work"
)

type Pixiv struct {
	conf    PixivConf
	client  *pixiv.Pixiv
	wm      *WorksMapper
	context *core.Context
}

type PixivConf struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
}

func (p *Pixiv) Status() core.ModuleStatus {
	if p.client == nil {
		return core.Status.Error(core.ErrUninitialized)
	}
	return core.Status.OK()
}

func (p *Pixiv) Init(id string, c *core.Context) error {
	ctx := context.Background()
	var err error
	p.wm, err = NewWorksMapper(ctx, c.Database(), c.DatabaseType())
	if err != nil {
		return err
	}
	if err := c.Config().Unmarshal("modules/pixiv", &p.conf); err != nil {
		return err
	}
	p.client = pixiv.New(p.conf.ClientID, p.conf.ClientSecret, p.conf.Username, p.conf.Password)
	if err := p.client.FetchToken(context.Background()); err != nil {
		p.client = nil
		return fmt.Errorf("Pixiv login failed: %v", err)
	}
	p.context = c
	c.Timeline().NewTopic(OriginPixivWork, DailyRankingKey)
	return nil
}

func (p *Pixiv) Start(c *core.Context) error {
	// first run
	if err := p.updateDaily(); err != nil {
		return err
	}
	sched := c.Scheduler()
	sched.Every(1).Days().Do(p.updateDaily, nil)
	return nil
}

func (p *Pixiv) Close() error {
	return nil
}

func (p *Pixiv) updateDaily() error {
	c := p.client
	now := time.Now()
	log.Infof("Updating /pixiv/ranking/daily")
	items, err := c.Ranking(pixiv.RANKING_CATEGORY_ILLUST, pixiv.RANKING_MODE_DAILY, 50, &now, 1)
	if err != nil {
		return err
	}
	wItems := make([]*pixiv.Item, 0, len(items))
	for i := range items {
		wItems = append(wItems, &items[i].Work)
	}
	if err := p.wm.InsertBulk(context.Background(), wItems, now); err != nil {
		return err
	}
	for _, item := range items {
		imageUrl := item.Work.ImageUrls[pixiv.SIZE_480x960]
		if imageUrl == "" {
			imageUrl = item.Work.ImageUrls[pixiv.SIZE_128x128]
		}
		if imageUrl == "" {
			imageUrl = item.Work.ImageUrls[pixiv.SIZE_50x50]
		}

		err = p.context.Timeline().Publish(DailyRankingKey, &timeline.Item{
			Caption:   item.Work.Title,
			Thumbnail: imageUrl,
			OriginKey: int64(item.Work.ID),
			Timestamp: now,
		})
		if err != nil {
			log.Errorf("Failed to publish pixiv daily: %v", err)
		}
	}
	return nil
}
