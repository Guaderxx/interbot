package amongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitMongo(uri string, db string) (*mongo.Database, error) {
	credentail := options.Credential{
		Username: "ader",
		Password: "123456",
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetAuth(credentail))
	if err != nil {
		return nil, err
	}

	return client.Database(db), nil
}

func Close(ctx context.Context, cli *mongo.Client) error {
	return cli.Disconnect(ctx)
}

func InitModels(db *mongo.Database) error {
	err := InitBotUserModel(db)
	if err != nil {
		return err
	}

	err = InitSessionStateModel(db)
	if err != nil {
		return err
	}

	return nil
}

func InitCollections(db *mongo.Database) map[string]*mongo.Collection {
	return map[string]*mongo.Collection{
		"botuser":  BotUserCollection(db),
		"groupmsg": GroupMsgCollection(db),
		"msgmap":   MsgMapCollection(db),
		"session":  SessionStateCollection(db),
		"topic":    TopicCollection(db),
	}
}

// FindOne 泛型查询单条文档
func FindOne[T any](
	ctx context.Context,
	collection *mongo.Collection,
	filter interface{},
	opts ...options.Lister[options.FindOneOptions],
) (*T, error) {
	var result T
	err := collection.FindOne(ctx, filter, opts...).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindMany 泛型查询多条文档
func FindMany[T any](
	ctx context.Context,
	collection *mongo.Collection,
	filter interface{},
	opts ...options.Lister[options.FindOptions],
) ([]T, error) {
	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
