package authentication

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMongoStore_CreateApp(t *testing.T) {
	store := NewMongoStore(client, database).(*mongoStore)
	defer client.Database(database).Drop(context.Background())

	info, err := store.CreateApp("test","testApp")
	require.NoError(t, err)
	require.Equal(t, "testApp", info.Description)
	require.Len(t, info.AppKey, appKeyLength)
	t.Log(info)
}

func TestMongoStore_GetApps(t *testing.T) {
	store := NewMongoStore(client, database).(*mongoStore)
	defer client.Database(database).Drop(context.Background())

	info, err := store.CreateApp("test","testApp")
	require.NoError(t, err)

	infos, err := store.GetApps()
	require.NoError(t, err)
	require.Len(t, infos, 1)

	require.Equal(t, info.AppId, infos[0].AppId)
	require.Equal(t, info.AppKey, infos[0].AppKey)
	require.Equal(t, info.Description, infos[0].Description)
	require.Equal(t, info.CreatedAt, infos[0].CreatedAt)
}

func TestMongoStore_GetAppKey(t *testing.T) {
	store := NewMongoStore(client, database).(*mongoStore)
	defer client.Database(database).Drop(context.Background())

	info, err := store.CreateApp("test", "testApp")
	require.NoError(t, err)

	key, err := store.GetAppKey(info.AppId)
	require.NoError(t, err)
	require.Equal(t, info.AppKey, key)
}
