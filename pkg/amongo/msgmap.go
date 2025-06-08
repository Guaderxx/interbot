package amongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InitMsgMapModel(db *mongo.Database) error {
	return nil
}

// MspMap 表示 msg 模型
type MsgMap struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	//
	UserID            int64 `bson:"user_id" json:"user_id"`                           // 对应的用户ID
	UserChatMessageID int   `bson:"user_chat_message_id" json:"user_chat_message_id"` // 对应的用户消息ID
	// 对应的
	GroupChatMessageID int `bson:"group_chat_message_id" json:"group_chat_message_id" validate:"required"` // TODO: need check

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// MsgMapCollection 返回用户集合的引用
func MsgMapCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("msgmaps")
}

func (s *MsgMap) InsertOne(ctx context.Context, db *mongo.Database) (*mongo.InsertOneResult, error) {
	col := MsgMapCollection(db)
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return col.InsertOne(ctx, s)
}
