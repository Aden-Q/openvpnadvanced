package config

import (
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

type AppConfig struct {
	AutoSubscribe bool
	UpdatePeriod  time.Duration
	CheckOpenVPN  bool
	LogLevel      string
}

var appConfig AppConfig

func LoadINIConfig(path string) error {
	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}
	appConfig.AutoSubscribe = cfg.Section("").Key("auto-subscribe").MustBool(false)
	appConfig.UpdatePeriod = cfg.Section("").Key("update-period").MustDuration(30 * time.Minute)
	appConfig.CheckOpenVPN = cfg.Section("").Key("check-openvpn").MustBool(true)
	appConfig.LogLevel = cfg.Section("").Key("log-level").MustString("info")
	return nil
}

func SaveINIConfig(path string) error {
	cfg := ini.Empty()
	cfg.Section("").Key("auto-subscribe").SetValue(fmt.Sprintf("%v", appConfig.AutoSubscribe))
	cfg.Section("").Key("update-period").SetValue(appConfig.UpdatePeriod.String())
	cfg.Section("").Key("check-openvpn").SetValue(fmt.Sprintf("%v", appConfig.CheckOpenVPN))
	cfg.Section("").Key("log-level").SetValue(appConfig.LogLevel)
	return cfg.SaveTo(path)
}

func GetConfig() AppConfig {
	return appConfig
}

func SetConfig(cfg AppConfig) {
	appConfig = cfg
}
