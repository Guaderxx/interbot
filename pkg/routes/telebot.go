package routes

import (
	"slices"
	"time"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/Guaderxx/interbot/pkg/core"
	"github.com/go-co-op/gocron/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/telebot.v4"
)

func Routes(c *core.Core) {
	c.Bot.Handle("/start", Wrapf(CmdStart, c))
	c.Bot.Handle(telebot.OnText, Wrapf(OnText, c))
	c.Bot.Handle(telebot.OnCallback, Wrapf(HandleCallback, c))
	c.Bot.Handle("/broadcast", Wrapf(CmdBroadcast, c))
	c.Bot.Handle("/clear", Wrapf(CmdClear, c))
}

// CmdBroadcast handle /broadcast command
func CmdBroadcast(co *core.Core, c telebot.Context) error {
	if c.Chat().ID != co.Config.AdminGroupID {
		return nil
	}

	if slices.Contains(co.Config.AdminUserIDs, c.Sender().ID) {
		return c.Reply("你没有权限执行此操作。")
	}

	// 检查是否是回复消息
	if c.Message().ReplyTo == nil {
		return c.Reply("这条指令需要回复一条消息，被回复的消息将被广播。")
	}

	// 安排广播任务
	_, err := co.Scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(time.Second*1))),
		gocron.NewTask(broadcast, co, c.Message().ReplyTo),
	)
	return err
}

func broadcast(co *core.Core, oriMsg *telebot.Message) error {
	logger := co.Logger.WithGroup("broadcast")
	// 实现广播逻辑
	var users []amongo.BotUser
	users, err := amongo.FindMany[amongo.BotUser](co.Ctx, co.Cols["botuser"], bson.D{})
	if err != nil {
		logger.Error("get users failed", "error", err)
		return err
	}
	for _, user := range users {
		chat, err := co.Bot.ChatByID(user.UserID)
		if err != nil {
			logger.Warn("get user chat failed", "user", user, "error", err)
			continue
		}
		_, err = co.Bot.Copy(chat, oriMsg)
		if err != nil {
			logger.Warn("copy message to user failed", "user", user, "error", err)
		}
	}
	return nil
}

// CmdClear handles the /clear command
func CmdClear(co *core.Core, c telebot.Context) error {
	logger := co.Logger.WithGroup("clear")
	if c.Chat().ID != co.Config.AdminGroupID {
		return nil
	}

	if !slices.Contains(co.Config.AdminUserIDs, c.Sender().ID) {
		return c.Reply("你没有权限执行此操作。")
	}

	topicID := c.Message().ThreadID
	logger.Info("delete topic", "chat", c.Chat().ID, "topic", topicID)

	err := co.Bot.DeleteTopic(c.Chat(), &telebot.Topic{
		ThreadID: topicID,
	})
	if err != nil {
		logger.Error("delete topic failed", "error", err)
		return err
	}
	if !co.Config.DeleteUserMessageOnClearCmd {
		return nil
	}

	user, err := amongo.FindBotUserByTopicID(co.Ctx, co.MDB, topicID)
	if err != nil {
		logger.Error("find user failed", "error", err)
		return err
	}

	msgmaps, err := amongo.FindMany[amongo.MsgMap](co.Ctx, co.Cols["msgmap"], bson.D{{"user_id", user.UserID}})
	if err != nil {
		logger.Error("find user messages failed", "error", err)
		return err
	}
	msgIDs := make([]int, 0, len(msgmaps))
	for _, msg := range msgmaps {
		msgIDs = append(msgIDs, msg.UserChatMessageID)
	}

	return co.Bot.DeleteMany(toMsgs(user.UserID, msgIDs))
}
