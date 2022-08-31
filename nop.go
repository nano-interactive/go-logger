package logging

var _ Error = &NopErrorLogger{}
var nopErrorLog = &NopErrorLogger{}

type NopErrorLogger struct{}

func (l *NopErrorLogger) Print(string, ...any) {}

