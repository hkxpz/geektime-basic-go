package config

var Config = WebookConfig{
	DB: DBConfig{DSN: "root:qwer1234.@tcp(120.25.240.120:63306)/webook"},
	Redis: RedisConfig{
		Addr:     "120.25.240.120:63879",
		DB:       0,
		Password: "qwer1234.",
	},
}
