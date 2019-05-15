package authentication

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/common"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/lt90s/goanalytics/utils"
	"net/http"
)

var userIdMissingError = utils.NewHttpError(http.StatusInternalServerError, 10001, "userID not found")

func getAccountInfoHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("userID")
		if userId == "" {
			c.Set("error", userIdMissingError)
		}
		info, err := adminStore.GetAccountInfo(userId)
		if err != nil {
			c.Set("error", err)
		} else {
			c.Set("data", info)
		}
	}
}

func GetAccountInfosHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		infos, err := adminStore.GetAccountInfos()
		if err != nil {
			c.Set("error", err)
		} else {
			c.Set("data", infos)
		}
	}
}

func isRoleValid(role string) bool {
	if role == "admin" || role == "operator" {
		return true
	}
	return false
}

func createAccountHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Name     string `json:"name"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}

		if data.Name == "" || len(data.Password) < 6 || !isRoleValid(data.Role) {
			c.Set("error", utils.ParamError)
			return
		}

		_, err = adminStore.CreateAccount(data.Name, data.Password, data.Role)

		if err != nil {
			c.Set("error", err)
		} else {
			c.Set("data", gin.H{})
		}

	}
}

func deleteAccountHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Id string `json:"id"`
		}
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}

		err = adminStore.DeleteAccount(data.Id)

		if err != nil {
			c.Set("error", err)
		} else {
			c.Set("data", gin.H{})
		}
	}
}

func getAppsHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		infos, err := adminStore.GetApps()
		if err != nil {
			c.Set("error", err)
		} else {
			c.Set("data", infos)
		}
	}
}

type createAppData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func createAppHandler(adminStore store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data createAppData
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}

		if data.Name == "" || data.Description == "" {
			c.Set("error", utils.ParamError)
			return
		}

		if len(data.Name) > 100 || len(data.Description) > 1000 {
			c.Set("error", utils.ParamError)
			return
		}
		info, err := adminStore.CreateApp(data.Name, data.Description)
		if err != nil {
			c.Set("error", utils.ParamError)
		} else {
			c.Set("data", info)
		}
	}
}

func deleteAppHandler(adminStore store, publisher pubsub.Publisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data common.DropDataRequest
		err := c.ShouldBindJSON(&data)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}
		err = adminStore.DeleteApp(data.AppId)
		if err != nil {
			c.Set("error", err)
		}
		publisher.Publish(common.GlobalEventDropData, &data)
	}
}
