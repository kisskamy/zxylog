# zxylog
一个轻量封装的log日志库, 基于对golang内置log的封装

# 功能
* 按照文件大小切割
* 按照日期切割
* 多实例管理

# 快速开始
```go
//创建日志
log := zxylog.NewZxyLog("./logs", "UserCenter")
log2 := zxylog.NewZxyLog("./logs", "Shop")

//也可以全局获取
//log = zxylog.NewLogManager().GetLogger("UserCenter")

//设置日志等级 ALL,DEBUG,INFO,WARN,ERROR,FATAL
log.SetLevel(zxylog.DEBUG)

//设置控制台打印前缀
log.SetConsolePrefix("UserCenter")

//设置是否控台打印, 默认true
log.SetConsole(true)

//例子
log.Debug("hello go")
log.Debugf("hello go %s", "!")
log.Debugln("hello", "go", "!!!")
log2.Debug("hello go")
log2.Debugf("hello go %s", "!")
log2.Debugln("hello", "go", "!!!")

```