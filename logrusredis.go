package logrusredis

import (
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"time"

	redis "github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

type RedisOption struct {
	Addr     string
	Password string
	DB       int
	Key      string
}

type LogWashFunc func(appName string, t time.Time, metadata logrus.Fields, caller *runtime.Frame, level logrus.Level, message string) []byte

type LogrusRedisHook struct {
	appName string
	option  *RedisOption
	rClient *redis.Client
	logWash LogWashFunc
}

func (h *LogrusRedisHook) Levels() []logrus.Level {
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

func (h *LogrusRedisHook) Fire(e *logrus.Entry) (err error) {
	bs := h.logWash(h.appName, e.Time, e.Data, e.Caller, e.Level, e.Message)
	if bs == nil {
		// ignore logs
		return nil
	}

	_, err = h.rClient.LPush(h.option.Key, bs).Result()
	return
}

func NewHook(appName string, option *RedisOption, logWashFunc LogWashFunc) (*LogrusRedisHook, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     option.Addr,
		Password: option.Password,
		DB:       option.DB,
	})
	if _, err := rdb.Ping().Result(); err != nil {
		return nil, err
	}
	if logWashFunc == nil {
		logWashFunc = DefaultLogWashFunc
	}
	return &LogrusRedisHook{
		appName: appName,
		rClient: rdb,
		option:  option,
		logWash: logWashFunc,
	}, nil
}

// -----------------------------------------------------------------------------

type DefaultLogS struct {
	Timestamp int64         `msgpack:"@timestamp"`
	MetaData  logrus.Fields `msgpack:"@metadata"` // map[string]interface{}
	Ip        string        `msgpack:"@ip"`
	Level     string        `msgpack:"@level"`
	Caller    string        `msgpack:"@caller"`
	Message   string        `msgpack:"@message"`
}

// https://stackoverflow.com/a/37382208/802815
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

var instanceIp string

func getIp() string {
	if instanceIp == "" {
		instanceIp = getOutboundIP()
	}
	return instanceIp
}

// DefaultLogWashFunc log filter and serialization
//   - use msgpack serialize log
//   - level > 'DEBUG'
//   - set current instance ip
func DefaultLogWashFunc(appName string, t time.Time, metadata logrus.Fields, caller *runtime.Frame, level logrus.Level, message string) []byte {
	if level > logrus.DebugLevel {
		return nil
	}

	caller_ := ""
	if caller != nil {
		caller_ = fmt.Sprintf("%s:%d", filepath.Base(caller.File), caller.Line)
	}

	l := DefaultLogS{
		Timestamp: t.UnixNano(),
		MetaData:  metadata,
		Ip:        getIp(),
		Level:     level.String(),
		Caller:    caller_,
		Message:   message,
	}

	bs, err := msgpack.Marshal(&l)
	if err != nil {
		return nil
	}

	return bs
}
