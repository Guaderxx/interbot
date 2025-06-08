package amongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitBotUserModel(db *mongo.Database) error {
	return EnsureBotUserIndexes(db)
}

// BotUser 表示机器人用户模型
type BotUser struct {
	ID                  bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Username            string        `bson:"username" json:"username" validate:"-"`
	UserID              int64         `bson:"id" json:"id" validate:"required"` // TODO: need check
	Balance             float64       `bson:"balance" json:"balance" validate:"gte=0"`
	ParentID            int64         `bson:"parent_id" json:"parent_id"`
	SourceID            string        `bson:"source_id" json:"source_id"`
	ConversationID      int           `bson:"conversation_id" json:"conversation_id"`
	DepositAddress      string        `bson:"depositAddress" json:"depositAddress"`
	Memo                string        `bson:"memo" json:"memo"`
	WithdrawalAddress   string        `bson:"withdrawalAddress" json:"withdrawalAddress"`
	TransactionPassword string        `bson:"transactionPassword" json:"transactionPassword"`
	Nickname            string        `bson:"nickname" json:"nickname" validate:"required"`
	IsPremium           bool          `bson:"isPremium" json:"isPremium"`
	CreatedAt           time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// BotUserCollection 返回用户集合的引用
func BotUserCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("botusers")
}

// EnsureBotUserIndexes 创建必要的索引
func EnsureBotUserIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 用户ID唯一索引
	userIDIndex := mongo.IndexModel{
		Keys:    bson.D{{"id", 1}},
		Options: options.Index().SetUnique(true),
	}

	// 2. 用户名索引
	usernameIndex := mongo.IndexModel{
		Keys:    bson.D{{"username", 1}},
		Options: options.Index().SetUnique(false),
	}

	// 3. 对话ID索引
	conversationIndex := mongo.IndexModel{
		Keys:    bson.D{{"conversation_id", 1}},
		Options: options.Index().SetUnique(false),
	}

	// 4. 提现地址索引
	withdrawalIndex := mongo.IndexModel{
		Keys:    bson.D{{"withdrawalAddress", 1}},
		Options: options.Index().SetUnique(false), // 根据需求可以设置为true
	}

	// 创建所有索引
	_, err := BotUserCollection(db).Indexes().CreateMany(ctx, []mongo.IndexModel{
		userIDIndex,
		usernameIndex,
		conversationIndex,
		withdrawalIndex,
	})

	return err
}

func FindBotUserByTGID(ctx context.Context, db *mongo.Database, userID int64) (*BotUser, error) {
	col := BotUserCollection(db)

	var user BotUser
	err := col.FindOne(ctx, bson.D{{"id", userID}}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func FindBotUserByTopicID(ctx context.Context, db *mongo.Database, topicID int) (*BotUser, error) {
	col := BotUserCollection(db)
	var user BotUser

	err := col.FindOne(ctx, bson.D{{"conversation_id", topicID}}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUser  if exist, load user, if not, create a new user
func GetUser(ctx context.Context, db *mongo.Database, user *BotUser) (*BotUser, error) {
	col := BotUserCollection(db)
	now := time.Now()

	err := col.FindOneAndUpdate(
		ctx,
		bson.D{{"id", user.UserID}},
		bson.D{{"$setOnInsert", bson.D{
			{"id", user.UserID},
			{"username", user.Username},
			{"nickname", user.Nickname},
			{"isPremium", user.IsPremium},
			{"createdAt", now},
			{"updatedAt", now},
		}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true),
	).Decode(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *BotUser) Update(ctx context.Context, db *mongo.Database, upt bson.D) (*mongo.UpdateResult, error) {
	col := BotUserCollection(db)
	upt = append(upt, bson.E{"updatedAt", time.Now()})
	return col.UpdateOne(ctx,
		bson.D{{"id", s.UserID}},
		bson.D{{"$set", upt}},
		options.UpdateOne().SetUpsert(true),
	)
}
