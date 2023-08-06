package config

var Config = WebookConfig{
	DB: DBConfig{DSN: "root:root@tcp(localhost:3306)/webook"},
	Redis: RedisConfig{
		Addr:     "120.25.240.120:63879",
		DB:       0,
		Password: "qwer1234.",
	},
}

//var Config = WebookConfig{
//	DB: DBConfig{
//		DSN: "root:root@tcp(localhost:13316)/webook",
//	},
//	Redis: RedisConfig{
//		Addr:     "localhost:6379",
//		Password: "",
//		DB:       1,
//	},
//}
