package common

const (
	GlobalEventDropData = "GlobalEventDropData"
)

type DropDataRequest struct {
	AppId string `json:appId`
}