package iniconfig

import (
	"fmt"
	"github.com/go-ini/ini"
	"sync"
)

var base *ini.Section
var iniFile *ini.File
var loadOnce sync.Once

func LoadINIConfig(filePath string) (err error) {
	loadOnce.Do(func() {
		iniFile, err = ini.Load(filePath)
		if err != nil {
			return
		}
		iniFile.BlockMode = false

		base, err = iniFile.GetSection("base")
		if err != nil {
			return
		}
	})
	return err
}

func NewAppConf(appType string, appId int32) (Config, error) {
	conf := Config{appType: appType, appId: appId}
	app, err := iniFile.GetSection(fmt.Sprintf("%s_%d", appType, appId))
	conf.app = app
	return conf, err
}

type Config struct {
	appType string
	appId   int32
	app     *ini.Section
}

func (c *Config) findKey(key string) *ini.Key {
	if base.HasKey(key) {
		return base.Key(key)
	}

	if c.app.HasKey(key) {
		return c.app.Key(key)
	}
	return nil
}

func (c *Config) AppId() int32    { return c.appId }
func (c *Config) AppType() string { return c.appType }

func (c *Config) Int32(key string) int32 {
	keyValue := c.findKey(key)
	if keyValue == nil {
		return 0
	}
	value, _ := keyValue.Int()
	return int32(value)
}

func (c *Config) Int64(key string) int64 {
	keyValue := c.findKey(key)
	if keyValue == nil {
		return 0
	}
	value, _ := keyValue.Int()
	return int64(value)
}

func (c *Config) String(key string) string {
	keyValue := c.findKey(key)
	if keyValue == nil {
		return ""
	}
	return keyValue.String()
}

func (c *Config) Bool(key string) bool {
	keyValue := c.findKey(key)
	if keyValue == nil {
		return false
	}
	value, _ := keyValue.Bool()
	return value
}

func BaseString(key string) string {
	if base.HasKey(key) {
		return base.Key(key).String()
	}
	return ""
}
