package config

var Config = WebookConfig{
	DB: DBConfig{DSN: "root:root@tcp(localhost:3306)/webook"},
	Redis: RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	},
}
