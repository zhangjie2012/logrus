package main

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

type FakeUser struct {
	ID   int64
	Name string
}

// 2000/s - 10000/s
func RandomAction(r *rand.Rand) (time.Duration, string) {
	// ms := r.Intn(400) + 100 // 100 - 500 Microsecond
	ms := 100
	actions := []string{
		"alley-oop",
		"assist",
		"behind",
		"blocking",
		"crossover",
		"cut",
		"dribble",
		"fake",
		"foul",
		"layup",
		"move",
		"pass",
		"shot",
		"slam dunk",
		"spin",
		"tie",
		"air ball",
	}
	return time.Duration(ms) * time.Microsecond, actions[ms%len(actions)]
}

func PlayGame(ctx context.Context, wg *sync.WaitGroup, user FakeUser) {
	defer wg.Done()

	logr := logrus.New()
	logr.SetOutput(ioutil.Discard)
	logr.SetLevel(logrus.TraceLevel)
	logr.SetReportCaller(true)

	var (
		key     string
		appName string
	)

	if user.ID < 6 {
		key = "logs_east"
		appName = "nba_east"
	} else {
		key = "logs_west"
		appName = "nba_weat"
	}
	option := logrusredis.Option{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Key:      key,
		AppName:  appName,
		Ip:       GetOutboundIP().String(),
	}

	hook, err := logrusredis.NewHook(&option, false)
	if err != nil {
		log.Fatal(err)
	}
	logr.AddHook(hook)

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	for {
		select {
		case <-ctx.Done():
			log.Println("receive done => ", user.Name)
			return
		default:
			d, action := RandomAction(r)
			logr.WithFields(logrus.Fields{"id": user.ID, "name": user.Name, "action": action}).Info(action)
			time.Sleep(d)
		}
	}
}

func main() {
	users := []FakeUser{
		{ID: 1, Name: "Lebron James"},
		{ID: 2, Name: "Darko Milicic"},
		{ID: 3, Name: "Carmelo Anthony"},
		{ID: 4, Name: "Chris Bosh"},
		{ID: 5, Name: "Dwyane Wade"},
		{ID: 6, Name: "Chris Kaman"},
		{ID: 7, Name: "Kirk Hinrich"},
		{ID: 8, Name: "TJ Ford"},
		{ID: 9, Name: "Mike Sweetney"},
		{ID: 10, Name: "Jarvis Hayes"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	for _, user := range users {
		wg.Add(1)
		go PlayGame(ctx, &wg, user)
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Ignore(syscall.SIGPIPE)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		cancel()
	}()

	wg.Wait()
}
