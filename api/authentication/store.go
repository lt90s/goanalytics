package authentication

import (
	"context"
	"errors"
	"github.com/lt90s/goanalytics/api/middlewares"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type store interface {
	CreateAccount(name, password, role string) (bool, error)
	DeleteAccount(id string) error
	AccountMatch(name, password string) (bool, middlewares.AccountMatchResult)
	GetAccountInfo(userID string) (info AccountInfo, err error)
	GetAccountInfos() (infos []AccountInfo, err error)

	GetApps() (infos []AppInfo, err error)
	GetAppIds() []string
	CreateApp(name, description string) (info AppInfo, err error)
	GetAppKey(appId string) (key string, err error)
	DeleteApp(appId string) error
}

type mongoStore struct {
	client   *mongo.Client
	database string
}

const (
	accountCollection = "accountCollection"
)

var (
	PasswordHashError    = errors.New("password hash error")
	NameAlreadyUsedError = errors.New("name already used")
	AccountNotExistError = errors.New("account not exist")
)

func NewMongoStore(client *mongo.Client, database string) store {

	ms := &mongoStore{
		client:   client,
		database: database,
	}

	ctx := context.Background()
	iv := ms.accountCollection().Indexes()

	unique := true
	_, err := iv.CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: &options.IndexOptions{Unique: &unique},
	})
	if err != nil && !strings.Contains(err.Error(), "IndexKeySpecsConflict") {
		panic(err)
	}
	return ms
}
