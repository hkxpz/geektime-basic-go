package logger

type Field struct {
	Key   string
	Value any
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Int[T int | int8 | int16 | int32 | int64](key string, val T) Field {
	return Field{Key: key, Value: val}
}

func Bool(key string, b bool) Field {
	return Field{Key: key, Value: b}
}

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Any(key string, val any) Field {
	return Field{Key: key, Value: val}
}
