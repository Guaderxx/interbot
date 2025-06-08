package core

import (
	"time"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"gopkg.in/telebot.v4"
)

func (c *Core) initBot() {
	pref := telebot.Settings{
		Token:  c.Config.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, ctx telebot.Context) {
			c.Logger.Error("bot error", "error", err)
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		c.Logger.Fatal("init bot failed", "error", err)
	}
	c.Logger.Info("bot initialized")
	c.Bot = bot
}

func (c *Core) midSession() {
	c.Bot.Use(func(hf telebot.HandlerFunc) telebot.HandlerFunc {
		return func(cc telebot.Context) error {
			if cc.Chat() == nil || cc.Sender() == nil || !cc.Message().Private() {
				c.Logger.Warn("empty chat or empty sender or not-private message, next")
				return hf(cc)
			}

			session, err := amongo.SetSession(c.Ctx, c.MDB, cc.Sender().ID)
			if err != nil {
				c.Logger.Warn("load session failed", "error", err)
				return cc.Reply("failed to load session, please try again later.")
			}
			cc.Set("session", session)
			return hf(cc)
		}
	})
}
