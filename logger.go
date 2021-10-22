package db

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type Logger interface {
	PRINT(v ...interface{})
	DEBUG(v ...interface{})
	WARN(v ...interface{})
	INFO(v ...interface{})
	ERROR(v ...interface{})
	FATAL(v ...interface{})
	PANIC(v ...interface{})
}

type logger struct {
	raw *logrus.Logger
}

func NewJSONLogger() Logger {
	raw := logrus.New()
	raw.SetFormatter(&logrus.JSONFormatter{})
	return &logger{
		raw: raw,
	}
}

func NewTextLogger() Logger {
	raw := logrus.New()
	raw.SetFormatter(&logrus.TextFormatter{})
	return &logger{
		raw: raw,
	}
}

func (l *logger) parseArgs(v []interface{}) (fields logrus.Fields, format string, args []interface{}, err error) {
	if len(v) == 0 || (len(v) == 1 && IsNil(v[0])) {
		return
	}
	if vv, ok := v[0].(string); ok {
		format = vv
		if len(v) > 1 {
			args = v[1:]
		}
	} else {
		indirectValue := reflect.Indirect(reflect.ValueOf(v[0]))
		switch indirectValue.Kind() {
		case reflect.Map, reflect.Struct:
			_ = JSONCopy(v[0], &fields)
		default:
			fields = make(logrus.Fields)
			fields["data"] = fmt.Sprintf("%v", v[0])
		}
		if len(v) > 1 {
			if vv, ok := v[1].(string); ok {
				format = vv
				if len(v) > 2 {
					args = v[2:]
				}
			}
		}
	}
	if len(fields) == 0 && format == "" {
		return fields, format, args, Errorf("unsupported parameter format")
	}
	return
}

func (l *logger) PRINT(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Printf(format, args...)
}

func (l *logger) DEBUG(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Debugf(format, args...)
}

func (l *logger) WARN(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Warnf(format, args...)
}

func (l *logger) INFO(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Infof(format, args...)
}

func (l *logger) ERROR(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Errorf(format, args...)
}

func (l *logger) FATAL(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Fatalf(format, args...)
}

func (l *logger) PANIC(v ...interface{}) {
	fields, format, args, err := l.parseArgs(v)
	if err != nil {
		return
	}
	logrus.WithFields(fields).Panicf(format, args...)
}
