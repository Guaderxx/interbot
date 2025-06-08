package config

import (
	"github.com/Guaderxx/interbot/pkg/alog"
)

type Config struct {
	AppName                     string  `mapstructure:"app_name"`
	BotToken                    string  `mapstructure:"bot_token"`
	WelcomeMessage              string  `mapstructure:"welcome_message"`
	AdminGroupID                int64   `mapstructure:"admin_group_id"`
	AdminUserIDs                []int64 `mapstructure:"admin_user_ids"`
	DeleteTopicAsForeverBan     bool    `mapstructure:"delete_topic_as_forever_ban"`
	DeleteUserMessageOnClearCmd bool    `mapstructure:"delete_user_message_on_clear_cmd"`
	DisableCaptcha              bool    `mapstructure:"disable_captcha"`
	MessageInterval             int64   `mapstructure:"message_interval"`

	Mongouri string
	Mdb      string

	Log alog.Options
}
