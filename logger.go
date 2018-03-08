package nano

import "github.com/sirupsen/logrus"

//Logger represents  the log interface
type Logger interface {
	Fatal(format ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
}

// Default logger
var logger Logger = logrus.New()

func init() {
	logger.(*logrus.Logger).SetLevel(logrus.DebugLevel)
}

// SetLogger rewrites the default logger
func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
}
