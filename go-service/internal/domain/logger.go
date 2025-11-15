package domain

// Field represents a structured logging value.
type Field struct {
	Key   string
	Value any
}

// Logger abstracts structured logging to keep domain free of concrete loggers.
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
}

// KV helper to build Field entries concisely.
func KV(key string, value any) Field {
	return Field{Key: key, Value: value}
}
