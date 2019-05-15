package authentication

import "go.mongodb.org/mongo-driver/bson/primitive"

type RegisterData struct {
	Name     string `json:"name"`
	Password string `json:"email"`
}

type AccountInfo struct {
	Id           string             `json:"id"`
	MongoId      primitive.ObjectID `json:"-" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Role         string             `json:"role" bson:"role"`
	PasswordHash []byte             `json:"-" bson:"passwordHash"`
}

var (
	ValidRoles = [...]string{"admin", "operator"}
)

type AppInfo struct {
	AppId       string             `json:"appId"`
	AppKey      string             `json:"appKey" bson:"appKey"`
	MongoId     primitive.ObjectID `json:"-" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   int64              `json:"createdAt" bson:"createdAt"`
}
