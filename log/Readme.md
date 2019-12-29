# 使用说明

- import "git.52retail.com/metro/metro/log"
- 默认为原生log输出
- 原生log输出可使用
```go
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
	
	log.SetFlags(olog.Ldate | olog.Ltime | olog.Lshortfile)

```
- 重新定义输出格式:
```go
func NewInfluxdbLogger(w io.Writer, measurement string, kv ...string) Logger

log.NewInfluxdbLogger(os.Stdout, "gotest")
```
 - 其中gotest为influxdb内的表名, kv为预设tag, 即所有日志都会带有

## 规则

 -  DEBUG 调试输出
 -	INFO  普通信息输出
 -	WARN  需要引起注意的输出
 - 	ERROR 错误输出(发生逻辑错误需解决的问题, 否则请使用WARN)


## 样例


```go
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Log("msg", "hello", "msg1", "hello1")
	log.Debug("msg", "hello", "msg1", "hello1")
	log.Info("msg", "hello", "msg1", "hello1")
	log.Warn("msg", "hello", "msg1", "hello1")
	log.Error("msg", "hello", "msg1", "hello1")
	l := log.WithHeader("AAA", "BBB")
	l.KVLog("msg", "hello", "msg1", "hello1")
	
	
//output
2019/07/04 18:15:19 gotest.go:442: [DEBUG]: msg hello msg1 hello1
2019/07/04 18:15:19 gotest.go:443: [DEBUG]: msg hello msg1 hello1
2019/07/04 18:15:19 gotest.go:444: [INFO]: msg hello msg1 hello1
2019/07/04 18:15:19 gotest.go:445: [WARN]: msg hello msg1 hello1
2019/07/04 18:15:19 gotest.go:446: [ERROR]: msg hello msg1 hello1
2019/07/04 18:15:19 gotest.go:448: [BBB]: msg hello msg1 hello1

```

- 打开influxdb格式输出
- 该模式使用WithHead函数自定义索引(慎用)
```go
	log.NewInfluxdbLogger(os.Stdout, "gotest")

//output	
gotest,level=DEBUG msg="msg hello msg1 hello1",caller="gotest.go:442" 1562236514671938896
gotest,level=DEBUG msg="hello",msg1="hello1",caller="gotest.go:443" 1562236514672017224
gotest,level=INFO msg="hello",msg1="hello1",caller="gotest.go:444" 1562236514672038321
gotest,level=WARN msg="hello",msg1="hello1",caller="gotest.go:445" 1562236514672053849
gotest,level=ERROR msg="hello",msg1="hello1",caller="gotest.go:446" 1562236514672082324
gotest,AAA=BBB msg="hello",msg1="hello1",caller="gotest.go:448" 1562236514672100931

```
