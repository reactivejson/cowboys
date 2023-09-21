package app

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"github.com/reactivejson/cowboys/internal/app"
	"github.com/reactivejson/cowboys/internal/domain"
	"github.com/reactivejson/cowboys/internal/game"
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

func SetupEnvConfig() *domain.MasterConfig {

	cfg := &domain.MasterConfig{}
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

func setupMasterService() setupFn {
	return func(c *Contx) (err error) {
		if c.masterService == nil {
			state := game.NewGame(c.cfg)
			c.masterService = app.NewMaster(c.cfg, state, c.log, c.redis)
			c.masterService.Run()
		}
		return nil
	}
}
