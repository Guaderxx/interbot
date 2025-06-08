package routes

import (
	"fmt"
	"slices"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/Guaderxx/interbot/pkg/core"
	"gopkg.in/telebot.v4"
)

// CmdStart 处理 /start 命令
func CmdStart(c *core.Core, ctx telebot.Context) error {
	logger := c.Logger.WithGroup("start")
	// Ignore non-private messages
	if !ctx.Message().Private() {
		logger.Warn("not private message", "chat", ctx.Chat().ID, "user", ctx.Sender().ID)
		return nil
	}

	au := parseUser(ctx.Sender())
	user, err := amongo.GetUser(c.Ctx, c.MDB, &au)
	if err != nil {
		logger.Warn("create user failed", "user", user, "error", err)
	} else {
		logger.Info("user created succeed", "user", user)
	}

	if slices.Contains(c.Config.AdminUserIDs, user.UserID) {
		logger.Info("admin user", "username", user.Nickname, "userid", user.ID)
		adminGroupChat, err := c.Bot.ChatByID(c.Config.AdminGroupID)
		if err != nil {
			logger.Error("get admin group failed", "error", err)
			return ctx.Reply(fmt.Sprintf("⚠️⚠️后台管理群组设置错误，请检查配置。⚠️⚠️\n你需要确保已经将机器人 @%s 邀请入管理群组并且给与了管理员权限。\n错误细节：%s\n", c.Config.AppName, err))
		}
		if adminGroupChat.Type == telebot.ChatSuperGroup || adminGroupChat.Type == telebot.ChatGroup {
			logger.Info("admin group", "group_name", adminGroupChat.Title, "group_id", adminGroupChat.ID)
			return ctx.Reply(fmt.Sprintf("你好管理员 %s(%d)\n\n欢迎使用 %s 机器人。\n\n 目前你的配置完全正确。可以在群组 <b> %s </b> 中使用机器人。", user.Nickname, user.ID, c.Config.AppName, adminGroupChat.Title), telebot.ModeHTML)
		}
	}

	logger.Info("user not admin", "user", user.Nickname)

	return ctx.Reply(fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a> 同学：\n\n%s", user.ID, user.Nickname, c.Config.WelcomeMessage), telebot.ModeHTML)
}
