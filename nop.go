package logger

var (
	_           Error = NopErrorLogger{}
	nopErrorLog       = NopErrorLogger{}
)

type NopErrorLogger struct{}

func (l NopErrorLogger) Print(string, ...any) {}
