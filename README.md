# logrusredis hook

a [logrus](https://github.com/sirupsen/logrus) hook print log to redis list(`RPUSH` command).

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
