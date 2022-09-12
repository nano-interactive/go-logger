package error_log

import (
	"fmt"
)

type MockErrorLogger struct {
	Buffer []string
}

func NewMockLogger() *MockErrorLogger {
	return &MockErrorLogger{
		Buffer: make([]string, 0, 1),
	}
}

func (l *MockErrorLogger) Print(format string, args ...any) {
	l.Buffer = append(l.Buffer, fmt.Sprintf(format, args...))
}
