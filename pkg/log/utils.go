package log

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
)

func RotateLogOutput(output string) io.Writer {

	return &lumberjack.Logger{
		Filename:   output,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
}
