package startup

import "geektime-basic-go/webook/pkg/logger"

func InitLog() logger.Logger {
	return logger.NewNoOpLogger()
}
