package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/zhangjie2012/logrusredis"
)

var (
	apiDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_log_report_durations",
			Help:       "http log report latency.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"api_name", "status"},
	)
)

func init() {
	prometheus.MustRegister(apiDurations)
}

type LogS struct {
	// fatal, error, warn, info, debug, trace
	// pass string not int, logrus.Level overwrite `UnmarshalText`
	Level    logrus.Level  `json:"level"`
	MetaData logrus.Fields `json:"metadata"`
	Msg      string        `json:"message"`
}

func main() {
	var (
		hostAddr      string = "localhost:8080"
		redisAddr     string = "localhost:6379"
		redisPassword string = ""
		redisDB       int    = 0
		redisKey      string = "logs_http"
	)
	flag.StringVar(&hostAddr, "hostaddr", hostAddr, "the server run address")
	flag.StringVar(&redisAddr, "redisaddr", redisAddr, "the redis address")
	flag.StringVar(&redisPassword, "redispassword", redisPassword, "the redis password")
	flag.IntVar(&redisDB, "redisdb", redisDB, "the redis db")
	flag.StringVar(&redisKey, "rediskey", redisKey, "the redis key")
	flag.Parse()

	// logrus init
	// logrus.SetOutput(ioutil.Discard)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(false)
	option := logrusredis.Option{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Key:      "logs_http",
		AppName:  "",
		Ip:       "",
	}
	logrusredis.NewHook(&option, false)
	hook, err := logrusredis.NewHook(&option, false)
	if err != nil {
		log.Fatal(err)
	}
	logrus.AddHook(hook)

	http.HandleFunc("/log", func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		status := ""
		defer func() {
			d := time.Since(start)
			apiDurations.WithLabelValues("/log", status).Observe(float64(d))
		}()

		defer req.Body.Close()
		bs, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("ioutil read body failure, %s", err)
			w.Write([]byte(err.Error()))
			status = "error"
			return
		}

		logS := LogS{}
		if err := json.Unmarshal(bs, &logS); err != nil {
			log.Printf("marshal json failure, %s", err)
			w.Write([]byte(err.Error()))
			status = "error"
			return
		}
		logrus.WithFields(logS.MetaData).Log(logS.Level, logS.Msg)

		w.Write([]byte("ok"))
		status = "ok"
	})

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	http.ListenAndServe(hostAddr, nil)
}
