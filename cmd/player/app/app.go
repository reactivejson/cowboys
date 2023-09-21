package app

import (
	"context"
)

/**
 * @author Mohamed-Aly Bou-Hanane
 * © 2023
 */

// setupFn is an options function receiving domain
type setupFn func(*Contx) error

// Setup is preparing the domain to be started
func Setup(ctx context.Context, cfg *Contx) error {
	defer runClosers(cfg.Closers) // close closers at the end

	setupFuncs := []setupFn{
		setupLog(),
		setupRedis(),
		setupPlayerService(),
	}
	return runSetupFncs(setupFuncs, cfg)
}
func runSetupFncs(fncs []setupFn, c *Contx) error {
	for _, f := range fncs {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func runClosers(fncs []func()) {
	for _, f := range fncs {
		f()
	}
}
