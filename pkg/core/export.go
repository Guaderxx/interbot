package core

import (
	"context"

	"github.com/Guaderxx/interbot/config"
	"github.com/Guaderxx/interbot/pkg/alog"
	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/go-co-op/gocron/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gopkg.in/telebot.v4"
)

type Core struct {
	Ctx    context.Context
	Config *config.Config
	Logger alog.ALogger

	Scheduler gocron.Scheduler
	Bot       *telebot.Bot
	MDB       *mongo.Database
	Cols      map[string]*mongo.Collection
}

func New(cfg config.Config) (*Core, error) {
	ctx := context.Background()
	c := &Core{
		Ctx:    ctx,
		Config: &cfg,
		Logger: alog.New(ctx, cfg.Log.Formatter, cfg.Log.Level, cfg.Log.AddSource),
	}

	mg, err := amongo.InitMongo(cfg.Mongouri, cfg.Mdb)
	if err != nil {
		return nil, err
	}
	c.Logger.Info("init mongo database succeed", "uri", cfg.Mongouri, "db", cfg.Mdb)

	err = amongo.InitModels(mg)
	if err != nil {
		return nil, err
	}
	c.Logger.Info("init mongo models succeed", "db", cfg.Mdb)

	c.MDB = mg

	c.Cols = amongo.InitCollections(mg)
	c.Logger.Info("init mongo collections succeed", "db", cfg.Mdb)

	c.initScheduer()
	c.initBot()
	c.midSession()

	return c, nil
}
