package logrusredis

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

var (
	appName = "logrusredishook"
	prod    = false
)

func TestNormal(t *testing.T) {
	option := RedisOption{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Key:      "logrusredis.hook",
	}
	hook, err := NewHook(appName, &option, nil)
	if err != nil {
		t.Fatal(err)
	}

	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	log.SetReportCaller(true)
	log.AddHook(hook)

	log.WithField("level", "error").Error("this is error log")
	log.WithField("level", "warn").Warn("this is warn log")
	log.WithField("level", "info").Info("this is info log")
	log.WithField("level", "debug").Debug("this is debug log")
}

type MyLogS struct {
	Timestamp int64         `msgpack:"@timestamp"`
	MetaData  logrus.Fields `msgpack:"@metadata"` // map[string]interface{}
	Level     string        `msgpack:"@level"`
	Message   string        `msgpack:"@message"`
}

func MyLogWashFunc(appName string, t time.Time, metadata logrus.Fields, caller *runtime.Frame, level logrus.Level, message string) []byte {
	// just error level
	if level == logrus.InfoLevel {
		return nil
	}

	// check metadata include or exclude
	_, ok := metadata["logid"]
	if !ok {
		return nil
	}

	l := MyLogS{
		Timestamp: t.Unix(), // to second prec
		MetaData:  metadata,
		Level:     level.String(),
		Message:   message,
	}

	fmt.Println(l)

	bs, err := msgpack.Marshal(&l)
	if err != nil {
		return nil
	}

	return bs
}

func TestCustomLogWash(t *testing.T) {
	option := RedisOption{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Key:      "logrusredis.hook",
	}
	hook, err := NewHook(appName, &option, nil)
	if err != nil {
		t.Fatal(err)
	}

	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	log.SetReportCaller(true)
	log.AddHook(hook)

	log.WithField("logid", "1").Info("any log")
	log.WithField("logid", "2").Info("any log")
	log.WithField("logid", "3").Info("any log")
	log.WithField("level", "debug").Debug("this is debug log") // will be ignore
}
