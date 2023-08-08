package main

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geektime-basic-go/week02/webook/config"
	"geektime-basic-go/week02/webook/internal/repository"
	"geektime-basic-go/week02/webook/internal/repository/dao"
	"geektime-basic-go/week02/webook/internal/service"
	"geektime-basic-go/week02/webook/internal/web"
	"geektime-basic-go/week02/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUser(server, db)

	log.Fatalln(server.Run(":8001"))
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
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}

			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	//usingSession(server)
	server.Use(new(middleware.JWTLoginMiddlewareBuilder).Build())
	return server
}

func usingSession(server *gin.Engine) {
	store := memstore.NewStore(
		[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
		[]byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"),
	)

	server.Use(sessions.Sessions("ssid", store))
	server.Use(new(middleware.LoginMiddlewareBuilder).IgnorePath("/users/login", "/users/signup").Build())
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}
