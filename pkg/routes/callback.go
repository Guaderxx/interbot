package routes

import (
	"strconv"
	"strings"
	"time"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/Guaderxx/interbot/pkg/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/telebot.v4"
)

func HandleCallback(co *core.Core, c telebot.Context) error {
	logger := co.Logger.WithGroup("callback")
	if c.Callback() == nil {
		logger.Warn("callback is nil", "chat", c.Chat().ID, "user", c.Sender().ID)
		return nil
	}

	data := c.Callback().Data
	if strings.HasPrefix(data, "vcode_") {
		return CallbackCaptcha(co, c, data)
	}
	logger.Warn("unknown callback data", "data", data, "chat", c.Chat().ID, "user", c.Sender().ID)

	return nil
}

// CallbackCaptcha  handles the captcha callback
func CallbackCaptcha(co *core.Core, c telebot.Context, data string) error {
	logger := co.Logger.WithGroup("callback_captcha")

	parts := strings.Split(c.Callback().Data, "_")
	code := parts[1]
	logger.Info("captcha code", "code", code, "chat", c.Chat().ID, "user", c.Sender().ID)
	userID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return c.RespondAlert("验证码格式错误，请重试。")
	}
	user := c.Callback().Sender

	session := c.Get("session").(amongo.SessionState)

	now := time.Now()
	if user.ID == userID {
		if code == session.Vcode {
			_, err = session.Update(co.Ctx, co.MDB, bson.D{{"isHuman", true}})
			if err != nil {
				logger.Warn("ignore update session error", "error", err, "session", session.ID.Hex())
			}
			c.RespondText("正确，欢迎。")
			c.RespondText(mentionHtml(user.ID, fullName(user)) + "， 欢迎。")
		} else {
			_, err = session.Update(co.Ctx, co.MDB, bson.D{{"errorTime", now}, {"isHuman", false}})
			if err != nil {
				logger.Warn("ignore update session error", "error", err, "session", session.ID.Hex())
			}
			_ = c.RespondText("~错误~，禁言2分钟")
		}
	}

	return c.Delete()
}
