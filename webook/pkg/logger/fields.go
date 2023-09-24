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
