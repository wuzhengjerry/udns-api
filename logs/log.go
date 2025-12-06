// Package log 日志模块
package log

import "github.com/sirupsen/logrus"

var (
	l *logrus.Logger
)

// C 全局缓存对象, 默认使用
func C() *logrus.Logger {
	if l == nil {
		panic("global log instance is nil")
	}
	return l
}

// SetGlobal 设置全局日志
func SetGlobal() error {
	l = Logger()
	return nil
}
