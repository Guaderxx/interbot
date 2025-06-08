package routes

import (
	"embed"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/Guaderxx/interbot/pkg/amongo"
	"github.com/Guaderxx/interbot/pkg/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/telebot.v4"
)

//go:embed assets/*
var imgFiles embed.FS

var (
	imgPrefix = "image_"
	imgSuffix = ".png"
	vcodes    = []string{
		"AdDMc",
		"AlwFu",
		"Asjxn",
		"AxMWi",
		"BOoXI",
		"CPeVb",
		"CYURB",
		"DgprN",
		"earlS",
		"EDtCV",
		"eFKLu",
		"eOYrT",
		"eSIdz",
		"EXplI",
		"EXYgv",
		"faALJ",
		"fdZBQ",
		"FGIUH",
		"Flpor",
		"Fyoek",
		"gaSuM",
		"Getag",
		"gmnrd",
		"GPsng",
		"GSaic",
		"gwvFf",
		"GYgia",
		"gYjsJ",
		"hMXpG",
		"htbQc",
		"ilEZh",
		"irfLA",
		"JagIC",
		"JDVFA",
		"JTEHj",
		"JwRMe",
		"jxWlK",
		"KeNEj",
		"KIWnb",
		"KoCie",
		"ktpyz",
		"kuwNB",
		"lBJcS",
		"lnvpd",
		"lRSUQ",
		"LXyJa",
		"mCgzT",
		"mdKPz",
		"MLAgo",
		"NdsKl",
		"nmVNa",
		"nWbet",
		"nZHuJ",
		"OePAM",
		"ofTnM",
		"OISXQ",
		"OPRqu",
		"oSWyY",
		"pBsaJ",
		"pEaLu",
		"pfcdr",
		"pLywk",
		"PSaHb",
		"PtquB",
		"QjMCf",
		"QLisC",
		"qplvW",
		"rAxZB",
		"RcHFg",
		"rDQbB",
		"rTQgs",
		"RxBmh",
		"smByj",
		"sMiXC",
		"TcZJO",
		"tpNnM",
		"tpohz",
		"tpouC",
		"TSRNI",
		"UfWET",
		"UMFrq",
		"UMIHV",
		"uwgCY",
		"UxuTa",
		"vpVOP",
		"wGdkx",
		"wrqeE",
		"WuJPj",
		"WVxEb",
		"XkrZa",
		"xLgWD",
		"XovdI",
		"XQBAT",
		"xrvtc",
		"XumDv",
		"YNlhA",
		"ZBHYs",
		"ZixjX",
		"zuZhH",
		"AkKgx",
	}
)

// toMsgs  converts a slice of message IDs to a slice of telebot.Editable messages.
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

func fullName(user *telebot.User) string {
	if strings.TrimSpace(user.LastName) != "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return user.FirstName
}

func mentionHtml(id int64, name string) string {
	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", id, name)
}

// TODO: `contact` is not sendable, so add the default avatar
func avatar() io.Reader {
	f, _ := imgFiles.Open("assets/avatar.png")
	return f
}

func randomEmbedImage() (filename, vcode string, err error) {
	files, err := imgFiles.ReadDir("assets")
	if err != nil {
		return "", "", err
	}
	fileName := files[rand.Intn(len(files))]
	vcode = strings.TrimSuffix(strings.TrimPrefix(fileName.Name(), imgPrefix), imgSuffix)
	return "assets/" + fileName.Name(), vcode, nil
}

func Wrapf(f1 func(*core.Core, telebot.Context) error, c *core.Core) telebot.HandlerFunc {
	return func(cc telebot.Context) error {
		return f1(c, cc)
	}
}

// GenerateRandomString   generates a random string of the specified length
func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// SendContactCard 发送带头像的联系卡片
func SendContactCard(co *core.Core, c telebot.Context, chatID int64, msgThreadID int, user *telebot.User) error {
	var text string
	if user.IsPremium {
		text = "🏆 高级会员"
	} else {
		text = "✈️ 普通会员"
	}
	// 1. 创建按钮
	buttons := [][]telebot.InlineButton{
		{
			{
				Text: text,
				// TODO: FIXME:
				URL: "https://github.com/MiHaKun/Telegram-interactive-bot",
			},
		},
	}

	if user.Username != "" {
		buttons = append(buttons, []telebot.InlineButton{
			{
				Text: "👤 直接联络",
				URL:  fmt.Sprintf("https://t.me/%s", user.Username),
			},
		})
	}

	// 2. 尝试获取用户头像
	photos, err := co.Bot.ProfilePhotosOf(user)
	if err != nil {
		co.Logger.Error("获取用户头像失败", "error", err)
	}

	// 3. 根据是否有头像选择发送方式
	if len(photos) > 0 {
		// 使用最大尺寸的头像
		largestPhoto := photos[0]
		for _, photo := range photos {
			if photo.FileSize > largestPhoto.FileSize {
				largestPhoto = photo
			}
		}

		// 发送带头像的卡片
		_, err := co.Bot.Send(
			telebot.ChatID(chatID),
			&largestPhoto,
			fmt.Sprintf(
				"👤 <a href=\"tg://user?id=%d\">%s</a>\n\n📱 %d\n\n🔗 @%s",
				user.ID,
				user.FirstName,
				user.ID,
				func() string {
					if user.Username == "" {
						return "无"
					}
					return user.Username
				}(),
			),
			&telebot.SendOptions{
				ReplyMarkup: &telebot.ReplyMarkup{InlineKeyboard: buttons},
				ParseMode:   telebot.ModeHTML,
				ThreadID:    msgThreadID, // 话题ID（telebot.v4 3.0+支持）
			},
		)
		co.Logger.Info("发送联系卡片", "chatID", chatID, "msgThreadID", msgThreadID, "error", err)
		return err

	} else {
		// TODO: Sendable interface not support contact now
		// 发送备用联系人卡片
		// contact := &telebot.Contact{
		// 	PhoneNumber: "11111", // 虚拟号码
		// 	FirstName:   user.FirstName,
		// 	LastName:    user.LastName,
		// }
		photo := &telebot.Photo{
			File: telebot.FromReader(avatar()),
		}

		_, err := co.Bot.Send(
			telebot.ChatID(chatID),
			photo,
			&telebot.SendOptions{
				ReplyMarkup: &telebot.ReplyMarkup{InlineKeyboard: buttons},
				ThreadID:    msgThreadID,
			},
		)
		co.Logger.Info("发送备用联系卡片", "chatID", chatID, "msgThreadID", msgThreadID, "error", err)
		return err
	}
}

// CheckHuman  check if user pass the captcha
func CheckHuman(co *core.Core, c telebot.Context) bool {
	logger := co.Logger.WithGroup("checkhuman")
	user := c.Sender()
	session := c.Get("session").(amongo.SessionState)

	if session.IsHuman {
		return true
	}
	// 两分钟内禁言
	if time.Since(session.ErrorTime) < time.Minute*2 {
		c.Reply("你已经被禁言,请稍后再尝试。", telebot.ModeHTML)
		return false
	}

	file, vcode, err := randomEmbedImage()
	if err != nil {
		logger.Error("get random image failed", "error", err)
		return false
	}

	// 2. 生成选项（与之前相同）
	codes := make([]string, 0, 8)
	for i := 0; i < 7; i++ {
		codes = append(codes, GenerateRandomString(5))
	}
	codes = append(codes, vcode)
	rand.Shuffle(len(codes), func(i, j int) { codes[i], codes[j] = codes[j], codes[i] })

	// 3. 读取嵌入的图片数据
	fileData, err := imgFiles.ReadFile(file)
	if err != nil {
		logger.Error("read embed image failed", "error", err)
		return false
	}

	// 4. 创建按钮
	buttons := make([][]telebot.InlineButton, 0)
	for i := 0; i < len(codes); i += 4 {
		row := make([]telebot.InlineButton, 0, 4)
		for j := i; j < i+4 && j < len(codes); j++ {
			row = append(row, telebot.InlineButton{
				Text: codes[j],
				Data: fmt.Sprintf("vcode_%s_%d", codes[j], user.ID),
			})
		}
		buttons = append(buttons, row)
	}

	// 5. 发送图片（通过字节流上传）
	photo := &telebot.Photo{
		File:    telebot.FromReader(strings.NewReader(string(fileData))),
		Caption: fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a> 请选择图片中的文字。", user.ID, user.FirstName),
	}

	err = c.Send(photo, &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{InlineKeyboard: buttons},
		ParseMode:   telebot.ModeHTML,
	})
	if err != nil {
		logger.Error("send image failed", "error", err)
		return false
	}

	_, _ = session.Update(co.Ctx, co.MDB, bson.D{{"vcode", vcode}})

	c.DeleteAfter(time.Minute * 1)

	return false
}

func parseUser(tu *telebot.User) amongo.BotUser {
	return amongo.BotUser{
		UserID:    tu.ID,
		Username:  tu.Username,
		Nickname:  fullName(tu),
		IsPremium: tu.IsPremium,
	}
}
