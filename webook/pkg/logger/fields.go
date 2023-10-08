package logger

type Field struct {
	Key   string
	Value any
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}

func Int64(key string, val int64) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Bool(key string, b bool) Field {
	return Field{
		Key:   key,
		Value: b,
	}
}
