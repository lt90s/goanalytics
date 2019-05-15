package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/metric/user"
	"github.com/lt90s/goanalytics/storage"
	"github.com/lt90s/goanalytics/utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type counterDescriptor struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}

type counterDescriptorData struct {
	Descriptors []counterDescriptor `json:"descriptors"`
}

func InstallCounterEndpoint(iRouter, oRouter *gin.RouterGroup, counter storage.Counter) {
	oRouter.POST("/counter", func(c *gin.Context) {
		var data counterDescriptorData
		err := c.ShouldBindJSON(&data)
		logrus.WithFields(logrus.Fields{"data": data}).Debug("counter request")
		if err != nil {
			c.Set("error", err)
			return
		}
		getCounters(c, data, counter)
	})

	oRouter.GET("/counter/trend", func(c *gin.Context) {
		getTrendData(c, counter)
	})

	oRouter.GET("/counter/customized", getCustomizedCountersHandler(counter))
	oRouter.POST("/counter/customized", addCustomizedCounterHandler(counter))
	oRouter.DELETE("/counter/customized", deleteCustomizedCounter(counter))

	iRouter.POST("/counter/customized", customizedCounterHandler(counter))
}

func addCustomizedCounterHandler(counter storage.Counter) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetString("appId")
		var data storage.CustomizedCounter
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}
		if !data.Valid() {
			c.Set("error", utils.ParamError)
			return
		}
		if err := counter.AddCustomizedCounter(appId, data); err != nil {
			c.Set("error", utils.ParamError)
			return
		}
	}
}

func getCustomizedCountersHandler(counter storage.Counter) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetString("appId")
		data, err := counter.GetCustomizedCounters(appId)
		if err != nil {
			c.Set("error", utils.ParamError)
		} else {
			c.Set("data", data)
		}
	}
}

func deleteCustomizedCounter(counter storage.Counter) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetString("appId")
		var tmp struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := c.ShouldBindJSON(&tmp); err != nil {
			c.Set("error", utils.ParamError)
			return
		}
		err := counter.DeleteCustomizedCounter(appId, tmp.Name, tmp.Type)
		if err != nil {
			c.Set("error", err)
		}
	}
}

func customizedCounterHandler(counter storage.Counter) gin.HandlerFunc {
	return func(c *gin.Context) {
		metaData, ok := middlewares.GetMetaData(c)
		if !ok {
			c.Set("error", utils.ParamError)
			return
		}
		var data struct {
			Name   string  `json"name"`
			Type   string  `json:"type"`
			Slot   string  `json:"slot"`
			Amount float64 `json:"amount"`
		}
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}
		customizedCounter, err := counter.GetCustomizedCounter(metaData.AppId, data.Name, data.Type)

		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}

		data.Name += storage.CustomizedCounterNameSuffix
		switch data.Type {
		case "simple":
			err = counter.AddSimpleCounter(metaData.AppId, data.Name, metaData.DateTimestamp, data.Amount)
		case "slot":
			err = utils.ParamError
			for _, slot := range customizedCounter.Slots {
				if slot == data.Slot {
					err = counter.AddSlotCounter(metaData.AppId, data.Name, data.Slot, metaData.DateTimestamp, data.Amount)
					break
				}
			}
		case "cpv":
			err = counter.AddSimpleCPVCounter(metaData.AppId, metaData.Channel, metaData.Platform,
				metaData.Version, data.Name, metaData.DateTimestamp, data.Amount)
		default:
			err = utils.ParamError
		}

		if err != nil {
			c.Set("error", err)
		}
	}
}

func getCounters(c *gin.Context, data counterDescriptorData, counter storage.Counter) {
	appId := c.GetString("appId")
	results := make(map[string]interface{})

	for _, descriptor := range data.Descriptors {
		var err error
		var result interface{}
		switch descriptor.Type {
		case "simple":
			result, err = getSimpleCounters(appId, descriptor, counter)
		case "slot":
			result, err = getSlotCounters(appId, descriptor, counter)
		case "cpv":
			result, err = getCpvCounters(appId, descriptor, counter)
		}
		if err != nil {
			c.Set("error", err)
			return
		}
		results[descriptor.Name] = result
	}
	c.Set("data", results)
}

func getSimpleCounters(appId string, descriptor counterDescriptor, counter storage.Counter) (data interface{}, err error) {
	switch descriptor.Operator {
	case "sum":
		data, err = counter.GetSimpleCounterSum(appId, descriptor.Name, descriptor.Start, descriptor.End)
	case "span":
		data, err = counter.GetSimpleCounterSpan(appId, descriptor.Name, descriptor.Start, descriptor.End)
	default:
		err = utils.NewHttpError(http.StatusBadRequest, http.StatusBadRequest,
			"Simple counter only support sum and span operator")
	}
	return
}

func getSlotCounters(appId string, descriptor counterDescriptor, counter storage.Counter) (data interface{}, err error) {
	switch descriptor.Operator {
	case "span":
		data, err = counter.GetSlotCounterSpan(appId, descriptor.Name, descriptor.Start, descriptor.End)
	}
	return
}

func getCpvCounters(appId string, descriptor counterDescriptor, counter storage.Counter) (data interface{}, err error) {
	ops := strings.Split(descriptor.Operator, "_")
	switch ops[0] {
	case "dateCPV":
		data, err = counter.GetSimpleCPVDateCPV(appId, descriptor.Name, descriptor.Start, descriptor.End)
	case "dateSum":
		data, err = counter.GetSimpleCPVSumDate(appId, descriptor.Name, descriptor.Start, descriptor.End)
	case "channelDateSum":
		if len(ops) != 2 {
			err = utils.NewHttpError(http.StatusBadRequest, http.StatusBadRequest, "missing channel in op")
			return
		}
		data, err = counter.GetSimpleCPVChannelSumDate(appId, descriptor.Name, ops[1], descriptor.Start, descriptor.End)
	}
	return
}

func getTrendData(c *gin.Context, counter storage.Counter) {
	appId := c.GetString("appId")
	delta7 := utils.TodayDiff(7).Unix()
	delta8 := utils.TodayDiff(8).Unix()
	delta14 := utils.TodayDiff(14).Unix()
	delta30 := utils.TodayDiff(30).Unix()
	delta31 := utils.TodayDiff(31).Unix()
	delta60 := utils.TodayDiff(60).Unix()
	yesterday := utils.TodayDiff(1).Unix()

	newUser7, err := counter.GetSimpleCPVSumTotal(appId, user.NewUserCPVCounter, delta7, yesterday)
	if err != nil {
		c.Set("error", err)
		return
	}
	newUser14, err := counter.GetSimpleCPVSumTotal(appId, user.NewUserCPVCounter, delta14, delta8)
	if err != nil {
		c.Set("error", err)
		return
	}

	activeUser7, err := counter.GetSimpleCPVSumTotal(appId, user.DailyActiveCPVCounter, delta7, yesterday)
	if err != nil {
		c.Set("error", err)
		return
	}
	activeUser14, err := counter.GetSimpleCPVSumTotal(appId, user.DailyActiveCPVCounter, delta14, delta8)
	if err != nil {
		c.Set("error", err)
		return
	}
	activeUser30, err := counter.GetSimpleCPVSumTotal(appId, user.DailyActiveCPVCounter, delta30, yesterday)
	if err != nil {
		c.Set("error", err)
		return
	}

	activeUser60, err := counter.GetSimpleCPVSumTotal(appId, user.DailyActiveCPVCounter, delta60, delta31)
	if err != nil {
		c.Set("error", err)
		return
	}

	retention7, err := averageNewUserRetention(counter, appId, delta7, yesterday)
	if err != nil {
		c.Set("error", err)
		return
	}
	retention14, err := averageNewUserRetention(counter, appId, delta14, delta8)
	if err != nil {
		c.Set("error", err)
		return
	}

	activeRetention7, err := averageActiveUserRetention(counter, appId, delta7, yesterday)
	if err != nil {
		c.Set("error", err)
		return
	}
	activeRetention14, err := averageActiveUserRetention(counter, appId, delta14, delta8)
	if err != nil {
		c.Set("error", err)
		return
	}

	nowTs := utils.NowTimestamp()
	totalUser, err := counter.GetSimpleCPVSumTotal(appId, user.NewUserCPVCounter, 0, nowTs)
	totalRegisteredUser, err := counter.GetSimpleCPVSumTotal(appId, user.NewRegisteredUserCPVCounter, 0, nowTs)

	c.Set("data", gin.H{
		"newUser7":            newUser7,
		"newUser14":           newUser14,
		"activeUser7":         activeUser7,
		"activeUser14":        activeUser14,
		"activeUser30":        activeUser30,
		"activeUser60":        activeUser60,
		"retention7":          retention7,
		"retention14":         retention14,
		"activeRetention7":    activeRetention7,
		"activeRetention14":   activeRetention14,
		"totalUser":           totalUser,
		"totalRegisteredUser": totalRegisteredUser,
	})
}

func averageNewUserRetention(counter storage.Counter, appId string, start, end int64) (float64, error) {
	days := (end - start) / (24 * 3600)

	retentionSpan, err := counter.GetSlotCounterSpan(appId, user.NewUserRetentionSlotCounter, start, end)
	if err != nil {
		return 0, err
	}
	newUserSpan, err := counter.GetSimpleCPVSumDate(appId, user.NewUserCPVCounter, start, end)

	var result float64
	for start <= end {
		counter, ok := retentionSpan[start]
		if !ok {
			continue
		}

		a, ok := counter["1"]
		if !ok {
			continue
		}
		b, ok := newUserSpan[start]
		if !ok || b == 0 {
			continue
		}
		result += a / b
		start += 24 * 3600
	}
	return result / float64(days), nil
}

func averageActiveUserRetention(counter storage.Counter, appId string, start, end int64) (float64, error) {
	days := (end - start) / (24 * 3600)

	retentionSpan, err := counter.GetSlotCounterSpan(appId, user.ActiveUserRetentionSlotCounter, start, end)
	if err != nil {
		return 0, err
	}
	newUserSpan, err := counter.GetSimpleCPVSumDate(appId, user.DailyActiveCPVCounter, start, end)

	var result float64
	for start <= end {
		counter, ok := retentionSpan[start]
		if !ok {
			continue
		}

		a, ok := counter["1"]
		if !ok {
			continue
		}
		b, ok := newUserSpan[start]
		if !ok || b == 0 {
			continue
		}
		result += a / b
		start += 24 * 3600
	}
	logrus.Debug(result, days)
	return result / float64(days), nil
}
