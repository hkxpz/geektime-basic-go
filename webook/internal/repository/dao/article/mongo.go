package article

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDBDAO struct {
	col     *mongo.Collection
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) DAO {
	return &mongoDBDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	index := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "author_id", Value: 1},
				{Key: "create_at", Value: 1},
			},
			Options: options.Index(),
		},
	}

	if _, err := db.Collection("articles").Indexes().CreateMany(ctx, index); err != nil {
		return err
	}
	_, err := db.Collection("published_articles").Indexes().CreateMany(ctx, index)
	return err
}

func (dao *mongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	art.ID = dao.node.Generate().Int64()
	now := time.Now().UnixMilli()
	art.CreateAt, art.UpdateAt = now, now
	_, err := dao.col.InsertOne(ctx, art)
	return art.ID, err
}

func (dao *mongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	art.UpdateAt = time.Now().UnixMilli()
	filter := bson.D{{Key: "id", Value: art.ID}, {Key: "author_id", Value: art.AuthorID}}
	sets := bson.M{"$set": art}
	res, err := dao.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func (dao *mongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var err error
	if art.ID > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		art.ID, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return art.ID, err
	}

	now := time.Now().UnixMilli()
	art.UpdateAt = now
	filter := bson.D{{Key: "id", Value: art.ID}, {Key: "author_id", Value: art.AuthorID}}
	sets := bson.M{"$set": art, "$setOnInsert": bson.M{"create_at": now}}
	_, err = dao.liveCol.UpdateOne(ctx, filter, sets, options.Update().SetUpsert(true))
	return art.ID, err
}

func (dao *mongoDBDAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	filter := bson.D{{Key: "id", Value: id}, {Key: "author_id", Value: author}}
	sets := bson.M{"$set": bson.M{"status": status, "update_at": time.Now().UnixMilli()}}
	res, err := dao.liveCol.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func (dao *mongoDBDAO) GetPubByID(ctx *gin.Context, id int64) (PublishedArticle, error) {
	res := dao.liveCol.FindOne(ctx, bson.M{"id": id})
	if res.Err() != nil {
		return PublishedArticle{}, res.Err()
	}
	var pub PublishedArticle
	err := res.Decode(&pub)
	return pub, err
}
