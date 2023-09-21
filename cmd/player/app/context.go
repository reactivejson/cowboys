package app

import (
	"github.com/go-redis/redis/v8"
	"github.com/reactivejson/cowboys/internal/app"
	"github.com/reactivejson/cowboys/internal/domain"
	"log"
)

/**
 * @author Mohamed-Aly Bou-Hanane
 * © 2023
 */

// Contx is application's content
type Contx struct {
	Closers       []func()
	log           *log.Logger
	cfg           *domain.PlayerConfig
	redis         *redis.Client
	playerService *app.Player
}

// NewContext instantiates new rte context object.
func NewContext(cfg *domain.PlayerConfig) *Contx {
	return &Contx{
		Closers: []func(){},
		cfg:     cfg,
	}
}
