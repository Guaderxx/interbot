package amongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitTopicModel(db *mongo.Database) error {
	return nil
}

// Topic 表示 group-->topic 模型
type Topic struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	// 对应的 tguserID
	ChatID         int64     `bson:"id" json:"id" validate:"required"`
	ConversationID int       `bson:"conversation_id" json:"conversation_id"`
	Status         string    `bson:"status" json:"status"`
	CreatedAt      time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time `bson:"updatedAt" json:"updatedAt"`
}

// TopicCollection 返回用户集合的引用
func TopicCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("topics")
}

func (s *Topic) Update(ctx context.Context, db *mongo.Database, upt bson.D) (*mongo.UpdateResult, error) {
	col := TopicCollection(db)
	upt = append(upt, bson.E{"updatedAt", time.Now()})
	return col.UpdateOne(ctx,
		bson.D{{"id", s.ChatID}},
		bson.D{{"$set", upt}},
		options.UpdateOne().SetUpsert(true),
	)
}

func (t *Topic) InsertOne(ctx context.Context, db *mongo.Database) (*mongo.InsertOneResult, error) {
	col := TopicCollection(db)
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	return col.InsertOne(ctx, t)
}

func GetTopic(ctx context.Context, db *mongo.Database, topicID int) (*Topic, error) {
	col := TopicCollection(db)
	var topic Topic
	err := col.FindOne(ctx, bson.D{{"conversation_id", topicID}}).Decode(&topic)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No topic found
		}
		return nil, err // Other error
	}
	return &topic, nil
}
