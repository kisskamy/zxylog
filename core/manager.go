package core

var logManager *LogManager

//日志实例管理
type LogManager struct {
	loggers map[string]*ZxyLog
}

//返回一个全局的日志实例管理者
func NewLogManager() *LogManager {
	return getLogManager()
}

//获取日志实例管理者
func getLogManager() *LogManager {
	if logManager == nil {
		return &LogManager{loggers: map[string]*ZxyLog{}}
	} else {
		return logManager
	}
}

//获取日志实例
func (logManager *LogManager) GetLogger(loggerName string) *ZxyLog {
	return getLogManager().loggers[loggerName]
}

//设置日志实例
func (logManager *LogManager) SetLogger(logFile *ZxyLog) {
	getLogManager().loggers[logFile.logFilename] = logFile
}
