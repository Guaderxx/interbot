package amongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitSessionStateModel(db *mongo.Database) error {
	return EnsureSessionIndexes(db)
}

// SessionState 表示用户session模型
type SessionState struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID         int64         `bson:"id" json:"id" validate:"required"`
	TopicID        int64         `bson:"topicID" json:"topicID"`
	CurrentGroupID string        `bson:"currentGroupID" json:"currentGroupID"` // 当前会话的 group ID
	Vcode          string        `bson:"vcode" json:"vcode"`                   // 验证码
	ErrorTime      time.Time     `bson:"errorTime" json:"errorTime"`           // 错误时间
	IsHuman        bool          `bson:"isHuman" json:"isHuman"`               // 是否是人类用户
	CreatedAt      time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// EnsureBotUserIndexes 创建必要的索引
func EnsureSessionIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userIDIndex := mongo.IndexModel{
		Keys:    bson.D{{"id", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := SessionStateCollection(db).Indexes().CreateOne(ctx, userIDIndex)

	return err
}

// BotUserCollection 返回用户集合的引用
func SessionStateCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("sessionstates")
}

func (s *SessionState) Update(ctx context.Context, db *mongo.Database, upt bson.D) (*mongo.UpdateResult, error) {
	col := SessionStateCollection(db)
	upt = append(upt, bson.E{"updatedAt", time.Now()})
	return col.UpdateOne(ctx,
		bson.D{{"id", s.UserID}},
		bson.D{{"$set", upt}},
		options.UpdateOne().SetUpsert(true),
	)
}

// SetSession   if exist, load session, if not, create a new session
func SetSession(ctx context.Context, db *mongo.Database, userID int64) (SessionState, error) {
	var s SessionState
	col := SessionStateCollection(db)
	now := time.Now()
	err := col.FindOneAndUpdate(ctx, bson.D{{"id", userID}}, bson.D{{"$setOnInsert", bson.D{
		{"id", userID},
		{"createdAt", now},
		{"updatedAt", now},
	}}}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&s)
	return s, err
}
