package logger

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING 
	ERROR
)

type Logger interface {
	Info(str string) 
	Warn(str string) 
	Error(str string)
	Debug(str string)
}
