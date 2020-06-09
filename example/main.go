package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/zhangjie2012/logrusredis"
)

// https://stackoverflow.com/a/37382208/802815
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func init() {
	logrus.SetOutput(ioutil.Discard)
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)

	option := logrusredis.Option{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Key:      "logs_test",
		AppName:  "test",
		Ip:       GetOutboundIP().String(),
	}
	hook, err := logrusredis.NewHook(&option, false)
	if err != nil {
		log.Fatal(err)
	}

	logrus.AddHook(hook)
}

func main() {
	logrus.Traceln("this is one trace line")
	logrus.Debugln("this is one debug line")
	logrus.Infoln("this is one info line")
	logrus.Warnln("this is one warn line")
	logrus.Errorln("this is one warn line")
	logrus.WithFields(logrus.Fields{"username": "JerryZhang"}).Infof("hello world")
}
