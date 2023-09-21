package app

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"github.com/reactivejson/cowboys/internal/app"
	"github.com/reactivejson/cowboys/internal/domain"
	"log"
)

/**
 * @author Mohamed-Aly Bou-Hanane
 * © 2023
 */

func setupLog() setupFn {
	return func(c *Contx) (err error) {
		if c.log == nil {
			c.log = log.Default()
		}
		return nil
	}
	//TODO proper logging setup
}

func SetupEnvConfig() *domain.PlayerConfig {

	cfg := &domain.PlayerConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		fmt.Errorf("could not parse config: %w", err)
	}
	return cfg
}

func setupRedis() setupFn {
	return func(c *Contx) (err error) {
		if c.redis == nil {
			options := &redis.Options{
				Addr: c.cfg.RedisAddr,
			}
			c.redis = redis.NewClient(options)
		}
		return nil
	}
}

func setupPlayerService() setupFn {
	return func(c *Contx) (err error) {
		if c.playerService == nil {
			c.playerService = app.NewPlayer(c.cfg, c.redis, c.log)
			c.playerService.Run()
		}
		return nil
	}
}
