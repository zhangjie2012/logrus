package logrusredis

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Option struct {
	Addr     string
	Password string
	DB       int
	Key      string
	AppName  string
	Ip       string
}

type LogrusRedisHook struct {
	option     *Option
	rClient    *redis.Client
	production bool
}

func (h *LogrusRedisHook) Levels() []logrus.Level {
	if h.production {
		return []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		}
	} else {
		return []logrus.Level{
			logrus.TraceLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		}
	}
}

type LogS struct {
	Timestamp int64         `json:"@timestamp"`
	AppName   string        `json:"@appname"`
	MetaData  logrus.Fields `json:"@metadata"` // map[string]interface{}
	Ip        string        `json:"@ip"`
	Level     string        `json:"@level"`
	Caller    string        `json:"@caller"`
	Msg       string        `json:"@msg"`
}

func (h *LogrusRedisHook) Fire(e *logrus.Entry) error {
	caller := fmt.Sprintf("%s:%d", filepath.Base(e.Caller.File), e.Caller.Line)
	log := LogS{
		Timestamp: e.Time.UnixNano(),
		AppName:   h.option.AppName,
		MetaData:  e.Data,
		Ip:        h.option.Ip,
		Level:     e.Level.String(),
		Caller:    caller,
		Msg:       e.Message,
	}

	logBs, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = h.rClient.LPush(h.option.Key, logBs).Result()
	if err != nil {
		return err
	}

	return nil
}

func NewHook(option *Option, production bool) (*LogrusRedisHook, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     option.Addr,
		Password: option.Password,
		DB:       option.DB,
	})
	if _, err := rdb.Ping().Result(); err != nil {
		return nil, err
	}
	return &LogrusRedisHook{
		rClient:    rdb,
		option:     option,
		production: production,
	}, nil
}
