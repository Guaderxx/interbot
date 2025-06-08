package amongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InitGroupMsgModel(db *mongo.Database) error {
	return nil
}

// GroupMsg 表示 media_group_message 模型
type GroupMsg struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	//
	ChatID int64 `bson:"conversation_id" json:"conversation_id"`
	// 对应的
	MessageID int       `bson:"id" json:"id" validate:"required"` // TODO: need check
	GroupID   string    `bson:"group_id" json:"group_id"`         // 群组ID
	IsHeader  bool      `bson:"is_header" json:"is_header"`       // 是否为头条消息
	Caption   string    `bson:"caption" json:"caption"`           // 消息标题
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// GroupMsgCollection 返回用户集合的引用
func GroupMsgCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("groupmsgs")
}

func (s *GroupMsg) InsertOne(ctx context.Context, db *mongo.Database) (*mongo.InsertOneResult, error) {
	col := GroupMsgCollection(db)
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	return col.InsertOne(ctx, s)
}
