package routes

import (
	"fmt"
	"time"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/Guaderxx/interbot/pkg/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/telebot.v4"
)

func OnText(co *core.Core, c telebot.Context) error {
	logger := co.Logger.WithGroup("ontext")
	// ignore empty message
	if c.Text() == "" {
		return c.Reply("请发送有效的文本消息。")
	}

	// private message, forward user message to admin group
	if c.Message().Private() {
		logger.Info("private msg", "chat", c.Chat().ID, "user", c.Sender().ID, "msg", c.Text())
		return ForwardMessageU2A(co, c)
	}

	// the chat is admin group, forward admin msg to user
	if c.Chat().ID == co.Config.AdminGroupID {
		logger.Info("admin group", "chat", c.Chat().ID, "user", c.Sender().ID, "msg", c.Text(), "topic", c.Message().ThreadID)
		return ForwardMessageA2U(co, c)
	}

	logger.Warn("not private msg and not in admin group",
		"chat", c.Chat(),
		"user", c.Sender(),
		"msg", c.Message(),
	)
	return nil
}

// ForwardMessageU2A  Must be private message
// As private chat, chatID is userID
// As group chat, chatID is groupID
func ForwardMessageU2A(co *core.Core, c telebot.Context) error {
	logger := co.Logger.WithGroup("u2a")

	if !co.Config.DisableCaptcha {
		logger.Warn("not disable captcha")
		if !CheckHuman(co, c) {
			logger.Warn("check human failed")
			return nil
		}
	}
	logger.Info("check human succeed")

	session := c.Get("session").(amongo.SessionState)
	// 设置了则检查
	if co.Config.MessageInterval > 0 {
		if session.UpdatedAt.Add(time.Duration(co.Config.MessageInterval) * time.Second).After(time.Now()) {
			return c.Reply("请不要频繁发送消息。", telebot.ModeHTML)
		}
	}
	logger.Info("check message interval succeed")

	au := parseUser(c.Sender())
	user, err := amongo.GetUser(co.Ctx, co.MDB, &au)
	if err != nil {
		return c.Reply("获取用户信息失败，请联系管理员。")
	}

	adminGroupID := co.Config.AdminGroupID
	adminGroupChat, err := co.Bot.ChatByID(co.Config.AdminGroupID)
	if err != nil {
		co.Logger.Error("get admin group failed", "error", err)
		return c.Reply("获取管理员群组失败，请联系管理员。")
	}

	topicID := user.ConversationID
	f, err := amongo.FindOne[amongo.Topic](co.Ctx, co.Cols["topic"], bson.M{"conversation_id": topicID})
	// 如果 topic 已经存在，且 status 为 closed，则不创建新的 topic
	if err == nil {
		if f.Status == "closed" {
			return c.Reply("客服已经关闭对话。如需联系，请利用其他途径联络客服回复和你的对话。", telebot.ModeHTML)
		}
	}
	// 如果 topic 不存在，则创建新的 topic
	if err != nil && topicID == 0 {
		topic, err := co.Bot.CreateTopic(adminGroupChat, &telebot.Topic{
			Name: fmt.Sprintf("%s|%d", user.Nickname, user.ID),
		})
		if err != nil {
			co.Logger.Error("create topic failed", "error", err, "user", user)
			return c.Reply("创建话题失败，请联系管理员。")
		}
		logger.Info("create topic succeed", "topic", topic)
		topicID = topic.ThreadID
		user.Update(co.Ctx, co.MDB, bson.D{{"conversation_id", topicID}})
		// user.ConversationID = topicID

		co.Bot.Send(adminGroupChat, fmt.Sprintf("新的用户 <a href=\"tg://user?id=%d\">%s</a> 开始了一个新的会话。", user.ID, user.Nickname), &telebot.SendOptions{
			ThreadID:  topicID,
			ParseMode: telebot.ModeHTML,
		})
		logger.Info("send new user message succeed, start a new topic", "user", user)

		newTopic := amongo.Topic{
			ConversationID: topicID,
			ChatID:         user.UserID,
			Status:         "opened",
		}
		newTopic.InsertOne(co.Ctx, co.MDB)

		err = SendContactCard(co, c, adminGroupID, topicID, c.Sender())
		if err != nil {
			logger.Warn("ignore send contact card failed", "error", err)
		} else {
			logger.Info("send contact card succeed")
		}
	}
	logger.Info("anyway, over the u2a")

	var opts *telebot.SendOptions = &telebot.SendOptions{
		ThreadID: topicID,
	}

	if c.Message().ReplyTo != nil {
		logger.Info("msg replyto is not nil")
		replyInUserChat := c.Message().ReplyTo.ID
		msgMap, err := amongo.FindOne[amongo.MsgMap](co.Ctx, co.Cols["msgmap"], bson.M{"user_chat_message_id": replyInUserChat})

		if err != nil {
			co.Logger.Error("get message map failed", "error", err)
		} else {
			// TODO:
			opts.ReplyParams = &telebot.ReplyParams{
				MessageID: msgMap.GroupChatMessageID,
			}
		}
	}

	var sent *telebot.Message
	logger.Info("msg replyto is nil")

	if c.Message().AlbumID != "" {
		logger.Info("msg is media group")
		msg := amongo.GroupMsg{
			ChatID:    c.Message().Chat.ID,
			MessageID: c.Message().ID,
			GroupID:   c.Message().AlbumID,
			IsHeader:  false,
			Caption:   c.Message().Caption,
		}
		msg.InsertOne(co.Ctx, co.MDB)

		if c.Message().AlbumID != session.CurrentGroupID {
			session.Update(co.Ctx, co.MDB, bson.D{{"current_group_id", session.CurrentGroupID}})

			co.SendMediaGroupLater(c.Message().Chat.ID, adminGroupID, msg.GroupID, "u2a", 5*time.Second)
		}
		return nil
	} else {
		logger.Info("not media group")
		chat, err := co.Bot.ChatByID(co.Config.AdminGroupID)
		if err != nil {
			co.Logger.Error("get admin group failed", "error", err)
			return c.Reply("获取管理员群组失败，请联系管理员。")
		}
		sent, err = co.Bot.Copy(chat, c.Message(), opts)
		if err != nil {
			co.Logger.Error("copy message failed", "error", err)
			return c.Reply("复制消息失败，请联系管理员。")
		}
	}

	msgMap := amongo.MsgMap{
		UserChatMessageID:  c.Message().ID,
		GroupChatMessageID: sent.ID,
		UserID:             user.UserID,
	}
	msgMap.InsertOne(co.Ctx, co.MDB)

	return nil
}

// ForwardMessageA2U   forward admin message to user
func ForwardMessageA2U(co *core.Core, c telebot.Context) error {
	logger := co.Logger.WithGroup("a2u")

	au := parseUser(c.Sender())
	amongo.GetUser(co.Ctx, co.MDB, &au)

	topicID := c.Message().ThreadID
	// equal 0 means this message is not in a thread
	// that means doesn't open the `topic` feature or in the `General topc`
	if topicID == 0 {
		logger.Warn("msg_thread_id equal 0")
		return nil
	}

	user, err := amongo.FindBotUserByTopicID(co.Ctx, co.MDB, topicID)
	if err != nil {
		logger.Error("get user failed", "error", err, "topicID", topicID)
		return c.Reply("获取用户信息失败，请联系管理员。")
	}
	session, err := amongo.SetSession(co.Ctx, co.MDB, user.UserID)
	if err != nil {
		logger.Error("get user session failed", "error", err)
		return c.Reply("get user session failed")
	}

	if c.Message().TopicClosed != nil {
		logger.Info("user topic closed")
		co.Bot.Send(telebot.ChatID(user.UserID), "对话已经结束。对方已经关闭了对话。你的留言将被忽略。")
		topic, err := amongo.GetTopic(co.Ctx, co.MDB, topicID)
		if err == nil {
			topic.Update(co.Ctx, co.MDB, bson.D{{"status", "closed"}})
		}

		return nil
	}

	if c.Message().TopicReopened != nil {
		logger.Info("topic reopened")
		co.Bot.Send(telebot.ChatID(user.UserID), "对方重新打开了对话。可以继续对话了。")
		topic, err := amongo.GetTopic(co.Ctx, co.MDB, topicID)
		if err == nil {
			topic.Update(co.Ctx, co.MDB, bson.D{{"status", "opened"}})
		}
		return nil
	}

	topic, err := amongo.GetTopic(co.Ctx, co.MDB, topicID)
	if err == nil {
		if topic.Status == "closed" {
			return c.Reply("对话已经结束。希望和对方联系，需要打开对话。")
		}
	}

	var opts *telebot.SendOptions = &telebot.SendOptions{}
	logger.Info("群组中，客服发了消息了")
	if c.Message().ReplyTo != nil && c.Message().ReplyTo.ID != 0 {
		replyInAdmin := c.Message().ReplyTo.ID
		msgmap, err := amongo.FindOne[amongo.MsgMap](co.Ctx, co.Cols["msgmap"], bson.D{{"group_chat_message_id", replyInAdmin}})
		if err == nil {
			opts.ReplyParams = &telebot.ReplyParams{
				MessageID: msgmap.UserChatMessageID,
			}
		}
	}

	var sent *telebot.Message

	if c.Message().AlbumID != "" {
		logger.Info("msg is media group")
		msg := amongo.GroupMsg{
			ChatID:    c.Message().Chat.ID,
			MessageID: c.Message().ID,
			GroupID:   c.Message().AlbumID,
			IsHeader:  false,
			Caption:   c.Message().Caption,
		}
		msg.InsertOne(co.Ctx, co.MDB)

		if c.Message().AlbumID != session.CurrentGroupID {
			session.Update(co.Ctx, co.MDB, bson.D{{"current_group_id", session.CurrentGroupID}})
			co.SendMediaGroupLater(c.Message().Chat.ID, user.UserID, msg.GroupID, "a2u", 5*time.Second)
		}
	} else {
		logger.Info("msg is not media group")
		chat, err := co.Bot.ChatByID(user.UserID)
		if err != nil {
			co.Logger.Error("get user chat failed", "error", err)
			return c.Reply("获取用户聊天失败，请联系管理员。")
		}
		sent, err = co.Bot.Copy(chat, c.Message(), opts)
		if err != nil {
			logger.Warn("copy message failed", "error", err)
		}
	}

	msgmap := amongo.MsgMap{
		GroupChatMessageID: c.Message().ID,
		UserChatMessageID:  sent.ID,
		UserID:             user.UserID,
	}
	msgmap.InsertOne(co.Ctx, co.MDB)

	return nil
}
