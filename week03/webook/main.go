package main

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geektime-basic-go/week03/webook/config"
	"geektime-basic-go/week03/webook/internal/repository"
	"geektime-basic-go/week03/webook/internal/repository/dao"
	"geektime-basic-go/week03/webook/internal/service"
	"geektime-basic-go/week03/webook/internal/web"
	"geektime-basic-go/week03/webook/internal/web/middleware"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	db := initDB()
	server := initWebServer()
	initUser(server, db)

	log.Fatalln(server.Run(":8081"))
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}

	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Jwt-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}

			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	server.Use(new(middleware.JWTLoginMiddlewareBuilder).Build())

	//redisStore := redis.NewClient(&redis.Options{Addr: config.Config.Redis.Addr, Password: config.Config.Redis.Password})
	//server.Use(ratelimit.NewBuilder(redisStore, time.Minute, 100).Build())
	return server
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}
