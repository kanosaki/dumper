package tumblr

import (
	"context"
	"github.com/k0kubun/pp"
	"github.com/kanosaki/dumper/core"
)

type Tumblr struct {
	conf   TumblrConf
	client *TumblrClient
}

func (t *Tumblr) Init(id string, c *core.Context) error {
	if err := c.Config().Unmarshal("modules/tumblr", &t.conf); err != nil {
		return err
	}
	conf := t.conf
	t.client = NewTumblrClient(
		conf.ConsumerKey, conf.ConsumerSecret,
		conf.Token, conf.TokenSecret,
		"", "http://api.tumblr.com")
	return nil
}

func (t *Tumblr) Start(c *core.Context) error {
	ctx := context.Background()
	dr, err := t.client.Dashboard(ctx, &DashboardParams{
		Limit: 10,
	})
	if err != nil {
		return err
	}
	pp.Println(dr.Posts)
	return nil
}

func (t *Tumblr) Close() error {
	return nil
}

func (t *Tumblr) Status() core.ModuleStatus {
	return core.Status.OK()
}

type TumblrConf struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	Token          string `yaml:"token"`
	TokenSecret    string `yaml:"token_secret"`
}
