package main

import (
	"fmt"
	"github.com/lt90s/goanalytics/api/authentication"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"math/rand"
	"time"
)

func generatePassword(length int) string {
	x := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*"
	password := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		index := rand.Intn(len(x))
		password[i] = x[index]
	}
	return string(password)
}

func main() {
	database := conf.GetConfString(conf.MongoDatabaseAdminKey)
	authStore := authentication.NewMongoStore(mongodb.DefaultClient, database)

	password := generatePassword(10)
	_, err := authStore.CreateAccount("root", password, "admin")
	if err != nil {
		fmt.Println("create root account failed, error: ", err.Error())
		return
	}

	fmt.Printf("new root account created, name=root password=%s", password)
}
