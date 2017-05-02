package core

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jasonlvhit/gocron"
	"github.com/kanosaki/dumper/common"
	"github.com/kanosaki/dumper/pkg/errors"
	"github.com/kanosaki/dumper/pkg/eslog"
	"github.com/kanosaki/dumper/shelf"
	"github.com/kanosaki/dumper/timeline"
	"github.com/kanosaki/dumper/web"
	elastic "gopkg.in/olivere/elastic.v5"
)

type Context struct {
	conf     *common.Config
	modules  map[string]Module
	storage  *shelf.Shelf
	es       *elastic.Client
	sched    *gocron.Scheduler
	rootLog  *eslog.Logger
	log      logrus.FieldLogger
	db       *sql.DB
	dbType   common.DBType
	timeline *timeline.Service
	web      *web.Server
}

func NewContext(confpath string) (*Context, error) {
	conf := common.NewConfig(confpath)
	c := &Context{
		conf:    conf,
		sched:   gocron.NewScheduler(),
		modules: make(map[string]Module),
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Context) init() error {
	c.log = logrus.StandardLogger()
	var coreConf common.CoreConfig
	if err := c.conf.Unmarshal("core", &coreConf); err != nil {
		return err
	}
	if coreConf.ElasticSearchURL != "" {
		es, err := elastic.NewClient(elastic.SetURL(coreConf.ElasticSearchURL))
		if err != nil {
			return err
		}
		c.es = es
		c.rootLog = eslog.New("dumper-context", es)
	} else {
		c.log.Warnf("Elasticsearch is not configured.")
	}
	if coreConf.StorageMetaDir != "" && coreConf.StorageDir != "" {
		c.storage = shelf.New(coreConf.StorageMetaDir, coreConf.StorageDir)
	} else {
		c.log.Warnf("Storage is not configured.")
	}
	var err error
	c.web, err = web.New(c.conf)
	if err != nil {
		return err
	}
	if coreConf.DBParam == "" || coreConf.DBType != "" {
		return fmt.Errorf("No database configuration found.")
	}
	dbType := coreConf.DBType
	if strings.HasPrefix(dbType, "sqlite") {
		c.dbType = common.SQLite
	} else if strings.HasPrefix(dbType, "mysql") {
		c.dbType = common.MySQL
	}
	p := coreConf.DBParam
	c.db, err = sql.Open(dbType, p)
	if err != nil {
		return err
	}
	tlStorage, err := timeline.NewStorage(dbType, p)
	if err != nil {
		return err
	}
	c.timeline = timeline.NewService(tlStorage)
	return nil
}

func (c *Context) AddModule(id string, mod Module) error {
	if err := mod.Init(id, c); err != nil {
		return err
	}
	c.modules[id] = mod
	return nil
}

func (c *Context) Start() error {
	errCh := make(chan error, len(c.modules))
	for _, m := range c.modules {
		go func(mod Module) {
			errCh <- mod.Start(c)
		}(m)
	}
	var errs []error
	for i := 0; i < len(c.modules); i++ {
		res := <-errCh
		if res != nil {
			errs = append(errs, res)
		}
	}
	if len(errs) != 0 {
		return errors.Multi(errs...)
	}
	return nil
}

func (c *Context) Close() error {
	errCh := make(chan error, len(c.modules))
	for _, m := range c.modules {
		go func(mod Module) {
			errCh <- m.Close()
		}(m)
	}
	var errs []error
	for i := 0; i < len(c.modules); i++ {
		res := <-errCh
		if res != nil {
			errs = append(errs, res)
		}
	}
	if len(errs) != 0 {
		return errors.Multi(errs...)
	}
	return nil
}

// Common components

func (c *Context) Storage() *shelf.Shelf {
	return c.storage
}

func (c *Context) Elastic() *elastic.Client {
	return c.es
}

func (c *Context) Scheduler() *gocron.Scheduler {
	return c.sched
}

func (c *Context) Config() *common.Config {
	return c.conf
}

func (c *Context) Database() *sql.DB {
	return c.db
}

func (c *Context) DatabaseType() common.DBType {
	return c.dbType
}

func (c *Context) Timeline() *timeline.Service {
	return c.timeline
}
