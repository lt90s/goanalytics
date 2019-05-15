package authentication

import (
	"context"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func (ms *mongoStore) accountCollection() *mongo.Collection {
	return ms.client.Database(ms.database).Collection(accountCollection)
}

func (ms *mongoStore) CreateAccount(name, password, role string) (ok bool, err error) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	_, err = ms.accountCollection().InsertOne(ctx, bson.M{
		"name":         name,
		"passwordHash": hash,
		"role":         role,
		"createdAt":    utils.NowTimestamp(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			err = NameAlreadyUsedError
		}
		return
	}
	ok = true
	return
}

func (ms *mongoStore) DeleteAccount(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = ms.accountCollection().DeleteOne(ctx, bson.M{"_id": _id})
	return err
}

func (ms *mongoStore) GetAccountInfo(userID string) (info AccountInfo, err error) {
	ctx := context.Background()

	_id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	result := ms.accountCollection().FindOne(ctx, bson.M{"_id": _id})
	err = result.Err()
	if err != nil {
		return
	}

	err = result.Decode(&info)
	info.Id = userID
	return
}

func (ms *mongoStore) AccountMatch(name, password string) (match bool, result middlewares.AccountMatchResult) {
	match = false
	filter := bson.M{
		"name": name,
	}
	option := &options.FindOneOptions{
		Projection: bson.M{
			"_id":          1,
			"role":         1,
			"passwordHash": 1,
		},
	}
	sr := ms.accountCollection().FindOne(context.Background(), filter, option)
	if sr.Err() != nil {
		return
	}
	var info AccountInfo
	err := sr.Decode(&info)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword(info.PasswordHash, []byte(password))
	if err != nil {
		return
	}

	match = true
	result = middlewares.AccountMatchResult{
		Id:   info.MongoId.Hex(),
		Role: info.Role,
	}
	return
}

func (ms *mongoStore) GetAccountInfos() (infos []AccountInfo, err error) {
	ctx := context.Background()

	cursor, err := ms.accountCollection().Find(ctx, bson.M{})
	if err != nil {
		return
	}

	var info AccountInfo
	for cursor.Next(ctx) {
		err = cursor.Decode(&info)
		info.Id = info.MongoId.Hex()
		infos = append(infos, info)
	}
	return
}
