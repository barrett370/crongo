package log

type Logger interface {
	Println(...any)
	Printf(string, ...any)
}

type NoopLogger struct{}

func (NoopLogger) Println(...any)        {}
func (NoopLogger) Printf(string, ...any) {}
