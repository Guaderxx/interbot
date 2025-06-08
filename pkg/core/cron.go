package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Guaderxx/interbot/pkg/alog"
	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/go-co-op/gocron/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/telebot.v4"
)

func (c *Core) initScheduer() {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		alog.Fatal("init scheduler failed", "error", err)
	}
	c.Scheduler = scheduler
}

// BanUserLater 延迟封禁用户
func (c *Core) BanUserLater(chat *telebot.Chat, user *telebot.User, banDuration time.Duration, delay time.Duration) (string, error) {
	jobName := "ban_" + generateJobID(chat.Recipient(), user.ID)

	_, err := c.Scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(delay))),
		gocron.NewTask(
			func() {
				// 在 telebot.v4 中设置封禁时长
				banUntil := time.Now().Add(banDuration).Unix()
				err := c.Bot.Ban(chat, &telebot.ChatMember{
					Rights:          telebot.Rights{CanSendMessages: false},
					RestrictedUntil: banUntil,
					User:            user,
				})

				if err != nil {
					alog.Error("Failed to ban user", "user", user.ID, "error", err)
				}
			},
		),
		gocron.WithName(jobName),
	)

	return jobName, err
}

// SendMediaGroupLater 延迟发送媒体组
func (c *Core) SendMediaGroupLater(chatID, targetID int64, mediaGroupID string, direction string, delay time.Duration) (string, error) {
	jobName := fmt.Sprintf("sendmediagroup_%d_%d_%s", chatID, targetID, direction)
	_, err := c.Scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(delay))),
		gocron.NewTask(
			func() {
				if err := c.sendMediaGroupLater(chatID, targetID, mediaGroupID, direction); err != nil {
					alog.Error("Failed to send media group message", "error", err)
				}
			},
		),
		gocron.WithName(jobName),
	)

	return jobName, err
}

func (c *Core) sendMediaGroupLater(fromChatID, targetID int64, mediaGroupID string, direction string) error {
	groupMsgs, err := amongo.FindMany[amongo.GroupMsg](c.Ctx, c.Cols["groupmsg"], bson.D{{"chat_id", fromChatID}, {"group_id", mediaGroupID}})
	if err != nil {
		return err
	}

	// 收集消息ID
	msgIDs := make([]int, 0, len(groupMsgs))
	for _, msg := range groupMsgs {
		msgIDs = append(msgIDs, msg.MessageID)
	}

	// 获取目标聊天
	targetChat, err := c.Bot.ChatByID(targetID)
	if err != nil {
		return fmt.Errorf("获取目标聊天失败: %v", err)
	}

	if direction == "u2a" {
		// 用户→群组（带话题ID）
		user, err := amongo.FindBotUserByTGID(c.Ctx, c.MDB, fromChatID)
		if err != nil {
			return fmt.Errorf("查询用户失败: %v", err)
		}

		sentMsgs, err := c.Bot.ForwardMany(targetChat, toMsgs(int64(fromChatID), msgIDs), &telebot.SendOptions{
			ThreadID: user.ConversationID,
		})
		if err != nil {
			return fmt.Errorf("转发消息失败: %v", err)
		}
		var msgMaps []amongo.MsgMap
		for i, sentMsg := range sentMsgs {
			msgMap := amongo.MsgMap{
				UserChatMessageID:  groupMsgs[i].MessageID,
				GroupChatMessageID: sentMsg.ID,
				UserID:             user.UserID,
			}
			msgMaps = append(msgMaps, msgMap)
		}
		_, err = c.Cols["msgmap"].InsertMany(c.Ctx, msgMaps)
		if err != nil {
			return fmt.Errorf("保存消息映射失败: %v", err)
		}
	} else {
		// 群组→用户
		sentMessages, err := c.Bot.ForwardMany(targetChat, toMsgs(fromChatID, msgIDs))
		if err != nil {
			return fmt.Errorf("转发消息失败: %v", err)
		}
		var msgMaps []amongo.MsgMap
		for i, sentMsg := range sentMessages {
			msgMap := amongo.MsgMap{
				UserChatMessageID:  groupMsgs[i].MessageID,
				GroupChatMessageID: sentMsg.ID,
				UserID:             targetID,
			}
			msgMaps = append(msgMaps, msgMap)
		}
		_, err = c.Cols["msgmap"].InsertMany(c.Ctx, msgMaps)
		if err != nil {
			return fmt.Errorf("保存消息映射失败: %v", err)
		}
	}
	return nil
}

// generateJobID 生成唯一的任务ID
func generateJobID(parts ...interface{}) string {
	var id string
	for _, part := range parts {
		id += "_" + toString(part)
	}
	return id[1:] // 去掉第一个下划线
}
func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

func toMsgs(chatID int64, msgs []int) []telebot.Editable {
	editableMsgs := make([]telebot.Editable, 0, len(msgs))
	for _, msgID := range msgs {
		editableMsgs = append(editableMsgs, &telebot.StoredMessage{
			ChatID:    chatID,
			MessageID: fmt.Sprintf("%d", msgID),
		})
	}
	return editableMsgs
}
