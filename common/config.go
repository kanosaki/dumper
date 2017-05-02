package common

import (
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v2"
)

type Config struct {
	base string
}

func NewConfig(base string) *Config {
	return &Config{
		base: base,
	}
}

func (c *Config) Unmarshal(key string, v interface{}) error {
	data, err := ioutil.ReadFile(path.Join(c.base, fmt.Sprintf("%s.yaml", key)))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

func (c *Config) Marshal(key string, v interface{}) error {
	out, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(c.base, key), out, 0600)
}

type CoreConfig struct {
	ElasticSearchURL string `yaml:"elasticsearch_url"`
	StorageMetaDir   string `yaml:"storage_meta_dir"`
	StorageDir       string `yaml:"storage_dir"`
	DBParam          string `yaml:"db_param"`
	DBType           string `yaml:"db_type"`
}
