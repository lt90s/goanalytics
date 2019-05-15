package conf

import "github.com/spf13/viper"

const (
	ServerAddr = "SERVER_ADDR"

	MongoDSNConfKey        = "MONGODB_DSN"
	MongoDatabasePrefixKey = "MONGODB_DATABASE_PREFIX"
	MongoDatabaseAdminKey  = "MONGODB_DATABASE_ADMIN"

	DebugConfKey = "GO_DEBUG"

	TimezoneConfKey = "Timezone"

	// JWT MIDDLEWARE CONFIG
	JWTRealmConfKey = "JWT_REAL_CONF_KEY"
	JWTKeyConfKey   = "JWT_KEY_CONF_key"

	TopicConfKey = "TOPIC"

	RMQGroupId   = "RMQ_GROUP_ID"
	RMQServeName = "RMQ_SERVER_NAME"

	KafkaBrokersConfKey                 = "KAFKA_BROKERS"
	KafkaGroupIdConfKey                 = "KAFKA_GROUP_ID"
	KafkaNumberOfHandleProcessorConfKey = "KAFKA_NUMBER_OF_HANDLER_PROCESSOR"
)

func init() {
	viper.SetDefault(ServerAddr, "127.0.0.1:5678")

	viper.SetDefault(MongoDSNConfKey, "mongodb://127.0.0.1:27017")
	viper.SetDefault(MongoDatabasePrefixKey, "goanalytics_")
	viper.SetDefault(MongoDatabaseAdminKey, "goanalytics_admin")
	viper.SetDefault(TimezoneConfKey, "Asia/Shanghai")

	// JWT Middleware Config defaults
	viper.SetDefault(JWTRealmConfKey, "example.com")
	viper.SetDefault(JWTKeyConfKey, "ba9e6e6fa65a1093e2daaa1ba20c416d7583041ccaaf6b274e6a89e5fca8f3c0")

	viper.SetDefault(DebugConfKey, true)

	viper.SetDefault(TopicConfKey, "GoAnalytics_Topic")
	viper.SetDefault(RMQGroupId, "goAnalytics_groupId")
	viper.SetDefault(RMQServeName, "localhost:9876")

	viper.SetDefault(KafkaBrokersConfKey, []string{"localhost:9092"})
	viper.SetDefault(KafkaGroupIdConfKey, "goAnalytics_kafka_group")
	viper.SetDefault(KafkaNumberOfHandleProcessorConfKey, 16)
	viper.AutomaticEnv()
}

func GetConfString(key string) string {
	return viper.GetString(key)
}

func GetConfInt64(key string) int64 {
	return viper.GetInt64(key)
}

func GetConfStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func GetConfByteSlice(key string) []byte {
	return []byte(viper.GetString(key))
}

func IsDebug() bool {
	return viper.GetBool(DebugConfKey)
}
