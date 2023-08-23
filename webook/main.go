package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/ioc"
)

func main() {
	server := initWebServer()
	log.Fatalln(server.Run(":8081"))
}

func initWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	db := ioc.InitDB()
	userCache := cache.NewRedisUserCache(cmdable)
	codeCache := cache.NewRedisCodeCache(cmdable)
	userHandler := initUserHandler(db, userCache, codeCache)
	middlewares := ioc.GinMiddlewares(cmdable)
	return ioc.InitWebServer(userHandler, middlewares...)
}

func initUserHandler(db *gorm.DB, userCache cache.UserCache, codeCache cache.CodeCache) *web.UserHandler {
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO, userCache)
	userSvc := service.NewUserService(userRepo)
	codeSvc := initCodeSvc(codeCache)
	return web.NewUserHandler(userSvc, codeSvc)
}

func initCodeSvc(codeCache cache.CodeCache) service.CodeService {
	smsSvc := ioc.InitSmsSvc()
	codeRepo := repository.NewCachedCodeRepository(codeCache)
	return service.NreSMSCodeService(smsSvc, codeRepo)
}
