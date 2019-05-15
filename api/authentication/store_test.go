package authentication

import (
	"context"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

var client *mongo.Client

const (
	database = "goanalytics_admin_test"
)

func init() {
	client = mongodb.NewMongoClient()
}

func TestMongoStore_CreateAccount(t *testing.T) {
	store := NewMongoStore(client, database).(*mongoStore)
	defer client.Database(database).Drop(context.Background())

	ok, err := store.CreateAccount("foo", "abcd", "admin")
	require.NoError(t, err)
	require.True(t, ok)

	result := store.accountCollection().FindOne(context.Background(), bson.M{"name": "foo"})
	require.NoError(t, result.Err())

	var info AccountInfo
	err = result.Decode(&info)

	require.NoError(t, err)

	require.Equal(t, "foo", info.Name)
	require.Equal(t, "admin", info.Role)

	_, err = store.CreateAccount("foo", "abcd", "admin")
	require.Equal(t, NameAlreadyUsedError, err)
}

func TestMongoStore_AccountMatch(t *testing.T) {
	store := NewMongoStore(client, database).(*mongoStore)
	defer client.Database(database).Drop(context.Background())

	ok, err := store.CreateAccount("foo", "abcd", "admin")
	require.NoError(t, err)
	require.True(t, ok)

	match, _ := store.AccountMatch("foo", "abcd")
	require.True(t, ok)

	require.True(t, match)

	match, _ = store.AccountMatch("foo", "abcde")
	require.False(t, match)

	match, _ = store.AccountMatch("bar", "abcd")
	require.False(t, match)
}

func BenchmarkMongoStore_AccountMatch(b *testing.B) {
	store := NewMongoStore(client, database).(*mongoStore)
	//defer client.Database(database).Drop(context.Background())

	store.CreateAccount("foo", "abcd", "admin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.AccountMatch("foo", "abcd")
	}
}
