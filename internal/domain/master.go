package domain

type MasterConfig struct {
	Port      string `envconfig:"PORT"               required:"false" default:":8080"`
	RedisAddr string `envconfig:"REDIS_ADDR"               required:"false" default:"redis:6379"`
	Players   int    `envconfig:"COMPETITORS"`
}
