package logger

import (
	"encoding/json"
	"testing"
	"time"
)

var (
	path       = "/Users/wanglei/Desktop/rtc-mcu-stream/go-src/qoe"
	level      = "debug"
	name       = "rtc-mcu-stream"
	maxSize    = 0
	maxBackend = 10
	maxAge     = 0
)

type message struct {
	AppId      string `json:"appId"`
	ClusterId  string `json:"clusterId"`
	EventCode  int    `json:"eventCode"`
	EventId    int    `json:"eventId"`
	EventMsg   string `json:"eventMsg"`
	ModuleId   string `json:"moduleId"`
	RoomId     string `json:"roomId"`
	ServerId   string `json:"serverId"`
	ServerUuid string `json:"serverUuid"`
	SessionId  string `json:"sessionId"`
	Time       int64  `json:"time"`
	Times      string `json:"times"`
	UserId     string `json:"userId"`
}

func TestInitLogObj(t *testing.T) {
	obj := InitLogJsonQoeBilling(path, name, level, maxBackend)
	msgLog := message{
		AppId:      "123",
		ClusterId:  "456",
		EventCode:  0,
		EventId:    0,
		EventMsg:   "11",
		ModuleId:   "22",
		RoomId:     "33",
		ServerId:   "44",
		ServerUuid: "555",
		SessionId:  "6666",
		Time:       0,
		UserId:     "7777",
	}
	for {
		msgLog.Times = time.Now().Format("2006-01-02 15:04:05")
		data, _ := json.Marshal(&msgLog)
		d := string(data)
		obj.Debug(d)
		obj.Info(d)
		obj.Infof(d)
		obj.Error(d)
		obj.Errorf(d)
		time.Sleep(1 * time.Minute)
		obj.Info("----------------")
	}
}
