package zxylog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LEVEL int //日志等级
type COLOR int //显示颜色
type STYLE int //显示样式

const (
	ALL   LEVEL = iota //所有日志
	DEBUG              //调试
	INFO               //信息
	WARN               //警告
	ERROR              //错误
	FATAL              //崩溃
)

const (
	CLR_BLACK   = COLOR(30) // 黑色
	CLR_RED     = COLOR(31) // 红色
	CLR_GREEN   = COLOR(32) // 绿色
	CLR_YELLOW  = COLOR(33) // 黄色
	CLR_BLUE    = COLOR(34) // 蓝色
	CLR_PURPLE  = COLOR(35) // 紫红色
	CLR_CYAN    = COLOR(36) // 青蓝色
	CLR_WHITE   = COLOR(37) // 白色
	CLR_DEFAULT = COLOR(39) // 默认
)

const (
	STYLE_DEFAULT   = STYLE(0) //终端默认设置
	STYLE_HIGHLIGHT = STYLE(1) //高亮显示
	SYTLE_UNDERLINE = STYLE(4) //使用下划线
	SYTLE_BLINK     = STYLE(5) //闪烁
	STYLE_INVERSE   = STYLE(7) //反白显示
	STYLE_INVISIBLE = STYLE(8) //不可见
)

const (
	logFlags       = log.Ldate | log.Lmicroseconds | log.Lshortfile //日志输出flag
	logConsoleFlag = 0                                              //console输出flag
	logMaxSize     = 512 * 1024 * 1024                              //单个日志文件最大大小
)

//日志文件类结构
type ZxyLog struct {
	sync.RWMutex                 //线程锁
	logDir           string      //日志存放目录
	logFilename      string      //日志基础名字
	timestamp        time.Time   //日志创建时的时间戳
	logFilePath      string      //当前日志路径
	logFile          *os.File    //当前日志文件实例
	logger           *log.Logger //当前日志操作实例
	logConsole       bool        //终端控制台显示控制，默认为true
	logConsolePrefix string      //终端控制台显示前缀
	logLevel         LEVEL       //日志级别
}

//日志系统初始化，需要提供存放目录和日志基础名字
func NewZxyLog(fileDir, fileName string) *ZxyLog {

	//目录修正
	dir := fileDir
	if fileDir[len(fileDir)-1] == '\\' || fileDir[len(fileDir)-1] == '/' {
		dir = fileDir[:len(fileDir)-1]
	}

	//初始化结构体
	var zxyLog = &ZxyLog{logDir: dir, logFilename: fileName, timestamp: time.Now(), logConsole: true, logLevel: ALL}
	zxyLog.Lock()
	defer zxyLog.Unlock()

	//创建文件
	fn := zxyLog.newLogFile()
	var err error
	zxyLog.logFile, err = os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err.Error())
	}

	//初始化日志
	zxyLog.logger = log.New(zxyLog.logFile, "", logFlags)
	log.SetFlags(logConsoleFlag)
	zxyLog.logFilePath = fn

	//启动文件监控模块
	go fileMonitor(zxyLog)

	//绑定管理
	getLogManager().SetLogger(zxyLog)
	return zxyLog
}

//设置终端控制台是否显示日志
func (zxyLog *ZxyLog) SetConsole(isConsole bool) {
	zxyLog.logConsole = isConsole
}

//设置终端控制台显示前缀
func (zxyLog *ZxyLog) SetConsolePrefix(prefix string) {
	zxyLog.logConsolePrefix = prefix
}

//设置日志记录级别，低于这个级别的日志将会被丢弃
func (zxyLog *ZxyLog) SetLevel(_level LEVEL) {
	zxyLog.logLevel = _level
}

//输出信息到终端控制台上
func (zxyLog *ZxyLog) console(ll LEVEL, args string) {
	if zxyLog.logConsole {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short

		now := time.Now()

		context := ""
		if len(zxyLog.logConsolePrefix) > 0 {
			context = fmt.Sprintf("[%04d/%02d/%02d_%02d:%02d:%02d.%06d] @%s #%s:%d %s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), time.Duration(now.Nanosecond())/(time.Microsecond), zxyLog.logConsolePrefix, file, line, args)
		} else {
			context = fmt.Sprintf("[%04d/%02d/%02d_%02d:%02d:%02d.%06d] #%s:%d %s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), time.Duration(now.Nanosecond())/(time.Microsecond), file, line, args)
		}

		switch ll {
		case DEBUG:
			log.Println(SprintColor(context, STYLE_DEFAULT, CLR_DEFAULT, CLR_DEFAULT))
		case INFO:
			log.Println(SprintColor(context, STYLE_DEFAULT, CLR_DEFAULT, CLR_DEFAULT))
		case WARN:
			log.Println(SprintColor(context, STYLE_DEFAULT, CLR_YELLOW, CLR_DEFAULT))
		case ERROR:
			log.Println(SprintColor(context, STYLE_HIGHLIGHT, CLR_RED, CLR_DEFAULT))
		case FATAL:
			log.Println(SprintColor(context, STYLE_HIGHLIGHT, CLR_PURPLE, CLR_DEFAULT))
		default:
			log.Println(SprintColor(context, STYLE_DEFAULT, CLR_DEFAULT, CLR_DEFAULT))
		}

	}
}

//输出Debug日志
func (zxyLog *ZxyLog) Debug(arg interface{}) {
	defer catchError()
	if zxyLog.logLevel <= DEBUG {
		context := fmt.Sprintf("DEBUG %s", fmt.Sprintln(arg))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(DEBUG, context)
	}
}

//输出Info日志
func (zxyLog *ZxyLog) Info(arg interface{}) {
	defer catchError()
	if zxyLog.logLevel <= INFO {
		context := fmt.Sprintf("INFO %s", fmt.Sprintln(arg))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(INFO, context)
	}
}

//输出Warn日志
func (zxyLog *ZxyLog) Warn(arg interface{}) {
	defer catchError()
	if zxyLog.logLevel <= WARN {
		context := fmt.Sprintf("WARN %s", fmt.Sprintln(arg))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(WARN, context)
	}
}

//输出Error日志
func (zxyLog *ZxyLog) Error(arg interface{}) {
	defer catchError()
	if zxyLog.logLevel <= ERROR {
		context := fmt.Sprintf("ERROR %s", fmt.Sprintln(arg))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(ERROR, context)
	}
}

//输出Fatal日志
func (zxyLog *ZxyLog) Fatal(arg interface{}) {
	defer catchError()
	if zxyLog.logLevel <= FATAL {
		context := fmt.Sprintf("FATAL %s", fmt.Sprintln(arg))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(FATAL, context)
	}
}

//输出Debug日志，支持格式化操作
func (zxyLog *ZxyLog) Debugf(format string, args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= DEBUG {
		context := fmt.Sprintf("DEBUG %s", fmt.Sprintf(format, args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(DEBUG, context)
	}
}

//输出Info日志，支持格式化操作
func (zxyLog *ZxyLog) Infof(format string, args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= INFO {
		context := fmt.Sprintf("INFO %s", fmt.Sprintf(format, args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(INFO, context)
	}
}

//输出Warn日志，支持格式化操作
func (zxyLog *ZxyLog) Warnf(format string, args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= WARN {
		context := fmt.Sprintf("WARN %s", fmt.Sprintf(format, args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(WARN, context)
	}
}

//输出Error日志，支持格式化操作
func (zxyLog *ZxyLog) Errorf(format string, args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= ERROR {
		context := fmt.Sprintf("ERROR %s", fmt.Sprintf(format, args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(ERROR, context)
	}
}

//输出Fatal日志，支持格式化操作
func (zxyLog *ZxyLog) Fatalf(format string, args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= FATAL {
		context := fmt.Sprintf("FATAL %s", fmt.Sprintf(format, args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(FATAL, context)
	}
}

//输出Debug日志
func (zxyLog *ZxyLog) Debugln(args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= DEBUG {
		context := fmt.Sprintf("DEBUG %s", fmt.Sprintln(args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(DEBUG, context)
	}
}

//输出Info日志
func (zxyLog *ZxyLog) Infoln(args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= INFO {
		context := fmt.Sprintf("INFO %s", fmt.Sprintln(args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(INFO, context)
	}
}

//输出Warn日志
func (zxyLog *ZxyLog) Warnln(args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= WARN {
		context := fmt.Sprintf("WARN %s", fmt.Sprintln(args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(WARN, context)
	}
}

//输出Error日志
func (zxyLog *ZxyLog) Errorln(args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= ERROR {
		context := fmt.Sprintf("ERROR %s", fmt.Sprintln(args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(ERROR, context)
	}
}

//输出Fatal日志
func (zxyLog *ZxyLog) Fatalln(args ...interface{}) {
	defer catchError()
	if zxyLog.logLevel <= FATAL {
		context := fmt.Sprintf("FATAL %s", fmt.Sprintln(args...))
		context = strings.TrimRight(context, "\n")
		zxyLog.RLock()
		defer zxyLog.RUnlock()
		_ = zxyLog.logger.Output(2, context)
		zxyLog.console(FATAL, context)
	}
}

//输出检查日志是否需要重新命名，比如说跨天，大小变化，或是文件已经不存在
func (zxyLog *ZxyLog) isMustRename() bool {

	//检查是否跨天
	if zxyLog.checkFileDate() {
		return true
	}

	//检查大小是否有变化
	if zxyLog.checkFileSize() {
		return true
	}

	//检查文件是否存在
	if zxyLog.checkFileExist() {
		return true
	} else {
		_ = os.MkdirAll(zxyLog.logDir, os.ModePerm)
	}

	return false
}

//检查文件日期是否已经跨天
func (zxyLog *ZxyLog) checkFileDate() bool {
	if time.Now().YearDay() != zxyLog.timestamp.YearDay() {
		return true
	}

	return false
}

//检查文件大小是否已经超过指定的大小
func (zxyLog *ZxyLog) checkFileSize() bool {

	fileInfo, err := os.Stat(zxyLog.logFilePath)
	if err != nil {
		return false
	}

	if fileInfo.Size() >= logMaxSize {
		return true
	}

	return false
}

//检查文件大小是否存在
func (zxyLog *ZxyLog) checkFileExist() bool {
	if !isFileExist(zxyLog.logFilePath) {
		return true
	}

	return false
}

//执行文件重命名操作
func (zxyLog *ZxyLog) rename() {
	zxyLog.timestamp = time.Now()
	fn := zxyLog.newLogFile()

	if zxyLog.logFile != nil {
		_ = zxyLog.logFile.Close()
	}

	zxyLog.logFile, _ = os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	zxyLog.logger = log.New(zxyLog.logFile, "", logFlags)
	zxyLog.logFilePath = fn
}

//获取新的日志文件的名称
func (zxyLog *ZxyLog) newLogFile() string {

	dir := fmt.Sprintf("%s/%04d-%02d-%02d/", zxyLog.logDir, zxyLog.timestamp.Year(), zxyLog.timestamp.Month(), zxyLog.timestamp.Day())
	_ = os.MkdirAll(dir, os.ModePerm)
	filename := fmt.Sprintf("%s/%s", dir, zxyLog.logFilename)

	fn := filename + ".log"
	if !isFileExist(fn) {
		return fn
	}

	n := 1
	for {
		fn = fmt.Sprintf("%s_%d.log", filename, n)
		if !isFileExist(fn) {
			break
		}
		n += 1
	}

	return fn
}

//判断文件是否存在
func isFileExist(file string) bool {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return false
	}
	if fileInfo.IsDir() {
		return false
	} else {
		return true
	}
}

//捕获程序错误，仅供内部使用
func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

//文件监控函数，循环检测文件是否需要重命名
func fileMonitor(zxyLog *ZxyLog) {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer.C:
			fileCheck(zxyLog)
		}
	}
}

//检查文件是否需要重命名，如果需要，那么执行重命名逻辑
func fileCheck(zxyLog *ZxyLog) {

	defer catchError()
	if zxyLog != nil && zxyLog.isMustRename() {
		zxyLog.Lock()
		defer zxyLog.Unlock()
		zxyLog.rename()
	}
}

/**
用颜色来显示字符串
 @param
	str					待显示的字符串
	s					显示样式
	fc					显示前景色
	bc					显示背景色
*/
func SprintColor(str string, s STYLE, fc, bc COLOR) string {

	_fc := int(fc)      //前景色
	_bc := int(bc) + 10 //背景色

	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, int(s), _bc, _fc, str, 0x1B)
}
