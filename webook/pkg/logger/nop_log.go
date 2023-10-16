package logger

type noOpLogger struct {
}

func NewNoOpLogger() Logger {
	return &noOpLogger{}
}

func (n *noOpLogger) Debug(msg string, args ...any) {}

func (n *noOpLogger) Info(msg string, args ...any) {}

func (n *noOpLogger) Warn(msg string, args ...any) {}

func (n *noOpLogger) Error(msg string, args ...any) {}

func (n *noOpLogger) Panic(msg string, args ...any) {}

func (n *noOpLogger) Fatal(msg string, args ...any) {}
