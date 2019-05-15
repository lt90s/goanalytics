package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/lt90s/goanalytics/storage"
	"github.com/lt90s/goanalytics/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type counter struct {
	databasePrefix string
	client         *mongo.Client
}

const (
	slotCounterCollectionNamePrefix = "slotCounter_"
	simpleCounterSlotName           = "defaultSlot"

	simpleCPVCounterCollectionNamePrefix = "simpleCPVCounter_"
	customizedCounterCollectionName      = "customizedCounterCollection"
)

func NewCounter(client *mongo.Client, databasePrefix string) storage.Counter {
	return &counter{
		databasePrefix: databasePrefix,
		client:         client,
	}
}

func (c *counter) database(appId string) *mongo.Database {
	return c.client.Database(c.databasePrefix + appId)
}

func (c *counter) DropAllCounter(appId string) {
	c.client.Database(c.databasePrefix + appId).Drop(context.Background())
}

func (c *counter) slotCounterCollection(appId string, counterName string) *mongo.Collection {
	return c.database(appId).Collection(slotCounterCollectionNamePrefix + counterName)
}

func (c *counter) customizedCounterCollection(appId string) *mongo.Collection {
	return c.database(appId).Collection(customizedCounterCollectionName)
}

func (c *counter) AddSimpleCounter(appId string, counterName string, dateTimestamp int64, amount float64) error {
	return c.AddSlotCounter(appId, counterName, simpleCounterSlotName, dateTimestamp, amount)
}

func (c *counter) GetSimpleCounterSpan(appId string, counterName string, startTimestamp, endTimestamp int64) (map[int64]float64, error) {
	slotCounters, err := c.GetSlotCounterSpan(appId, counterName, startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}

	counters := make(map[int64]float64)
	for dateTimestamp, slotCounter := range slotCounters {
		counters[dateTimestamp] = slotCounter[simpleCounterSlotName]
	}
	return counters, nil
}

func (c *counter) GetSimpleCounterSum(appId string, counterName string, startTimestamp, endTimestamp int64) (float64, error) {
	sums, err := c.GetSlotCounterSum(appId, counterName, startTimestamp, endTimestamp, []string{simpleCounterSlotName})
	if err != nil {
		return 0, err
	}
	return sums[simpleCounterSlotName], nil
}

func (c *counter) AddSlotCounter(appId string, counterName, slotName string, dateTimestamp int64, amount float64) error {
	ctx := context.Background()
	filter := bson.M{"date": dateTimestamp}
	update := bson.M{"$inc": bson.M{"counter." + slotName: amount}}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := c.slotCounterCollection(appId, counterName).UpdateOne(ctx, filter, update, option)
	return err
}

func (c *counter) GetSlotCounterPartialSlotSum(appId string, counterName string, date int64, slots []string) float64 {
	ctx := context.Background()
	filter := bson.M{"date": date}
	result := c.slotCounterCollection(appId, counterName).FindOne(ctx, filter)
	var tmp struct {
		Counter map[string]float64 `bson:"counter"`
	}
	err := result.Decode(&tmp)
	if err != nil {
		return 0
	}

	var sum float64
	for _, slot := range slots {
		sum += tmp.Counter[slot]
	}
	return sum
}

func (c *counter) GetSlotCounterSpan(appId string, counterName string, start, end int64) (slotCounters storage.SlotCounters, err error) {
	ctx := context.Background()
	filter := bson.M{"date": bson.M{"$gte": start, "$lte": end}}
	cursor, err := c.slotCounterCollection(appId, counterName).Find(ctx, filter)
	if err != nil {
		return
	}

	var opaque struct {
		Date    int64       `bson:"date"`
		Counter interface{} `bson:"counter"`
	}
	slotCounters = make(storage.SlotCounters)
	for cursor.Next(ctx) {
		if err = cursor.Decode(&opaque); err != nil {
			return
		}
		if err = cursor.Err(); err != nil {
			return
		}
		type KV struct {
			Key   string
			Value int
		}

		counterArray, ok := opaque.Counter.(primitive.D)
		if !ok {
			err = errors.New("counter is not type of primitive.D")
			return
		}
		slotCounter := make(storage.SlotCounter)

		for _, e := range counterArray {
			slot := e.Key
			count := e.Value.(float64)
			slotCounter[slot] = count
		}
		slotCounters[opaque.Date] = slotCounter
	}
	return
}

func (c *counter) GetSlotCounterSum(appId string, counterName string, start, end int64, slots []string) (sums map[string]float64, err error) {
	if len(slots) == 0 {
		return
	}

	group := bson.M{
		"_id": nil,
	}
	for _, slot := range slots {
		group[slot] = bson.M{
			"$sum": fmt.Sprintf("$counter.%s", slot),
		}
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{
					"$gte": start,
					"$lte": end,
				},
			},
		},
		{
			"$group": group,
		},
	}
	ctx := context.Background()
	cursor, err := c.slotCounterCollection(appId, counterName).Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	data := make(map[string]interface{})
	if cursor.Next(ctx) {
		err = cursor.Decode(&data)
		if err != nil {
			return
		}

		sums = make(map[string]float64)
		for key, value := range data {
			if key == "_id" {
				continue
			}
			sums[key] = value.(float64)
		}
	}
	return
}

func (c *counter) simpleCPVCounterCollection(appId, counterName string) *mongo.Collection {
	return c.database(appId).Collection(simpleCPVCounterCollectionNamePrefix + counterName)
}

func (c *counter) AddSimpleCPVCounter(appId string, channel, platform, version, counterName string, dateTimestamp int64, amount float64) error {
	ctx := context.Background()
	filter := bson.M{
		"date":     dateTimestamp,
		"channel":  channel,
		"platform": platform,
		"version":  version,
	}
	update := bson.M{
		"$inc": bson.M{"counter": amount},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := c.simpleCPVCounterCollection(appId, counterName).UpdateOne(ctx, filter, update, option)
	return err
}

func (c *counter) GetSimpleCPVSumTotal(appId, counterName string, start, end int64) (float64, error) {
	ctx := context.Background()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{
					"$gte": start,
					"$lte": end,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"sum": bson.M{"$sum": "$counter"},
			},
		},
	}

	cursor, err := c.simpleCPVCounterCollection(appId, counterName).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var tmp struct {
		Sum float64 `bson:"sum"`
	}
	if cursor.Next(ctx) {
		err := cursor.Decode(&tmp)
		if err != nil {
			return 0, err
		}
	}
	return tmp.Sum, nil
}

func (c *counter) getSimpleCPVPartialSumDate(appId, counterName, partial, partialMatch string, start, end int64) (map[int64]float64, error) {
	ctx := context.Background()
	match := bson.M{
		"date": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	if partial == "C" {
		match["channel"] = partialMatch
	} else if partial == "P" {
		match["platform"] = partialMatch
	} else if partial == "V" {
		match["version"] = partialMatch
	}
	pipeline := []bson.M{
		{
			"$match": match,
		},
		{
			"$group": bson.M{
				"_id": "$date",
				"sum": bson.M{"$sum": "$counter"},
			},
		},
	}

	cursor, err := c.simpleCPVCounterCollection(appId, counterName).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var tmp struct {
		Date int64   `bson:"_id"`
		Sum  float64 `bson:"sum"`
	}
	sums := make(map[int64]float64)
	for cursor.Next(ctx) {
		err := cursor.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		sums[tmp.Date] = tmp.Sum
	}
	return sums, nil
}

func (c *counter) GetSimpleCPVChannelSumDate(appId, counterName, channel string, start, end int64) (map[int64]float64, error) {
	return c.getSimpleCPVPartialSumDate(appId, counterName, "C", channel, start, end)
}

func (c *counter) GetSimpleCPVSumDate(appId, counterName string, start, end int64) (map[int64]float64, error) {
	return c.getSimpleCPVPartialSumDate(appId, counterName, "", "", start, end)
}

func (c *counter) GetSimpleCPVDateCPV(appId, counterName string, start, end int64) (map[string]map[int64]map[string]float64, error) {
	ctx := context.Background()
	baselines := []string{"channel", "platform", "version"}
	dateCPV := make(map[string]map[int64]map[string]float64)

	entry := logrus.WithFields(logrus.Fields{"appId": appId, "counterName": counterName, "start": start, "end": end})
	entry.Debug("[GetSimpleCPVDateCPV] start")
	for _, baseline := range baselines {
		dateCPV[baseline] = make(map[int64]map[string]float64)
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"date": bson.M{"$lte": end, "$gte": start},
				},
			},
			{
				"$group": bson.M{
					"_id":        bson.M{"date": "$date", "metric": "$" + baseline},
					"counterSum": bson.M{"$sum": "$counter"},
				},
			},
		}
		cursor, err := c.simpleCPVCounterCollection(appId, counterName).Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		for cursor.Next(ctx) {
			var tmp struct {
				Id struct {
					Date   int64  `bson:"date"`
					Metric string `bson:"metric"`
				} `bson:"_id"`
				CounterSum float64 `bson:"counterSum"`
			}
			if err = cursor.Decode(&tmp); err != nil {
				logrus.Debug("[GetSimpleCPVDateCPV] decode error")
				return nil, err
			}
			if _, ok := dateCPV[baseline][tmp.Id.Date]; !ok {
				dateCPV[baseline][tmp.Id.Date] = make(map[string]float64)
			}
			dateCPV[baseline][tmp.Id.Date][tmp.Id.Metric] = tmp.CounterSum
		}
	}
	entry.WithFields(logrus.Fields{"dateCPV": dateCPV}).Debug("[GetSimpleCPVDateCPV] result")
	return dateCPV, nil
}

func (c *counter) GetCustomizedCounter(appId, name, type_ string) (data storage.CustomizedCounter, err error) {
	ctx := context.Background()
	filter := bson.M{
		"name": name + storage.CustomizedCounterNameSuffix,
		"type": type_,
	}
	sr := c.customizedCounterCollection(appId).FindOne(ctx, filter)

	if err = sr.Err(); err != nil {
		return
	}

	err = sr.Decode(&data)
	return
}

func (c *counter) AddCustomizedCounter(appId string, data storage.CustomizedCounter) error {
	data.Name = data.Name + storage.CustomizedCounterNameSuffix
	_, err := c.customizedCounterCollection(appId).InsertOne(context.Background(), data)
	return err
}

func (c *counter) GetCustomizedCounters(appId string) (counters []storage.CustomizedCounter, err error) {
	ctx := context.Background()
	cursor, err := c.customizedCounterCollection(appId).Find(ctx, bson.M{})
	if err != nil {
		return
	}

	var tmp storage.CustomizedCounter
	todayTimestamp := utils.TodayTimestamp()
	yesterdayTimestamp := todayTimestamp - 24*3600

	for cursor.Next(ctx) {
		err = cursor.Decode(&tmp)
		if err != nil {
			return
		}
		switch tmp.Type {
		case "simple":
			tmp.TodayCount, _ = c.GetSimpleCounterSum(appId, tmp.Name, todayTimestamp, todayTimestamp)
			tmp.YesterdayCount, _ = c.GetSimpleCounterSum(appId, tmp.Name, yesterdayTimestamp, yesterdayTimestamp)
		case "slot":
			tmp.TodayCount = c.GetSlotCounterPartialSlotSum(appId, tmp.Name, todayTimestamp, tmp.Slots)
			tmp.YesterdayCount = c.GetSlotCounterPartialSlotSum(appId, tmp.Name, yesterdayTimestamp, tmp.Slots)
		case "cpv":
			tmp.TodayCount, _ = c.GetSimpleCPVSumTotal(appId, tmp.Name, todayTimestamp, todayTimestamp)
			tmp.YesterdayCount, _ = c.GetSimpleCPVSumTotal(appId, tmp.Name, yesterdayTimestamp, todayTimestamp)
		}
		tmp.Name = strings.TrimSuffix(tmp.Name, storage.CustomizedCounterNameSuffix)
		counters = append(counters, tmp)
	}

	return
}

func (c *counter) DeleteCustomizedCounter(appId, name, type_ string) error {
	ctx := context.Background()

	name = name + storage.CustomizedCounterNameSuffix
	_, err := c.customizedCounterCollection(appId).DeleteOne(ctx, bson.M{
		"name": name,
		"type": type_})

	if err != nil {
		return err
	}
	switch type_ {
	case "simple", "slot":
		_, err = c.slotCounterCollection(appId, name).DeleteMany(ctx, bson.M{})
	case "cpv":
		_, err = c.simpleCPVCounterCollection(appId, name).DeleteMany(ctx, bson.M{})
	}

	return err
}
