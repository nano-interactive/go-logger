package logging

import (
	"fmt"
)

var _ Error = &mockErrorLogger{}

type mockErrorLogger struct {
	Buffer []string
}

func newMockLogger() *mockErrorLogger {
	return &mockErrorLogger{
		Buffer: make([]string, 0, 1),
	}
}

func (l *mockErrorLogger) Print(format string, args ...any) {
	l.Buffer = append(l.Buffer, fmt.Sprintf(format, args...))
}
