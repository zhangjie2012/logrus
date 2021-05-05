# logrusredis hook migrate to https://github.com/zhangjie2012/logrus-hook

a [logrus](https://github.com/sirupsen/logrus) hook print log to redis list(`RPUSH` command).

## Background

logrus is my common structure log package for Go. It print log to standard output or files(by hook),
so I can collect log through filebeat/journalbeat to ELK.

But in some scenario:

- ELK for me is too heavyweight. for example, [grafana/loki](https://github.com/grafana/loki) is more suitable for me;
- I want file special logs for data statistics, like `logrus.Fields` marked `{"logid": "xxx"}`;

So, one side filter log and insert log to message queue, the other side, consume log from message queue, insert to MongoDB/MySQL/Loki/...

Redis `list` can be used as a simple message queue by `RPUSH` `LPOP`.
Other message queue tool (Kafka) can also as one choose, but I like redis(it simple deploy, simple use).


## Install

```sh
go get github.com/zhangjie2012/logrusredis-hook
```

## Usage

``` go
  option := RedisOption{
	  Addr:     "localhost:6379",
	  Password: "",
	  DB:       0,
	  Key:      "logrusredis.hook",
  }
  hook, _ := NewHook(appName, &option, nil)
  logrus.AddHook(hook)
```

*Sometime*, We want customize inserted redis struct, you can customize a `LogWashFunc`, checkout `logrusredis_test.go` file.
