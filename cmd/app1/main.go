package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/Guaderxx/interbot/config"
	"github.com/Guaderxx/interbot/pkg/alog"
	"github.com/Guaderxx/interbot/pkg/core"
	"github.com/Guaderxx/interbot/pkg/routes"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	cfg         config.Config
	showVersion bool
)

func initConfig() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			alog.Fatal("get user home dir error", "error", err)
		}
		cfgFile = filepath.Join(home, ".config", "beforsrs.toml")
	}

	err := load(cfgFile, &cfg)

	if err != nil {
		alog.Fatal("load config error", "error", err)
	}
	alog.Info("init config succeed", "config", cfg)
}

func init() {
	// TODO: FIXME: change to current dir
	flag.StringVar(&cfgFile, "config", "$HOME/.config/tgbot.toml", "config file")
	flag.Parse()
	initConfig()
	viper.SetDefault("author", "Guaderxx <guaderxx@gmail.com>")
	// viper.SetDefault("license", "GPL")
}

func main() {
	c, err := core.New(cfg)
	if err != nil {
		alog.Fatal("init core failed", "error", err)
	}
	c.Scheduler.Start()
	c.Logger.Info("start scheduler")
	defer c.Scheduler.Shutdown()

	if _, err := c.Scheduler.NewJob(
		gocron.DurationJob(1*time.Second),
		gocron.NewTask(func() {}),
	); err != nil {
		c.Logger.Fatal("start scheduler failed", "error", err)
	}

	routes.Routes(c)

	c.Logger.Info("bot started")
	c.Bot.Start()
}

// load the config file from the location.
func load(loc string, config *config.Config) error {
	dir, file := filepath.Split(loc)
	viper.SetConfigName(file)
	fileType := filepath.Ext(file)
	viper.SetConfigType(fileType[1:])

	viper.AddConfigPath(dir)
	// optionally look for config in the working directory
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			alog.Fatal("load config", "location", loc, "error", "not found")
		} else {
			alog.Fatal("load config", "location", loc, "error", err)
		}
	}

	err = viper.Unmarshal(config)
	return err
}
