package main

import (
	"context"
	"fmt"
	"github.com/reactivejson/cowboys/cmd/master/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/**
 * @author Mohamed-Aly Bou-Hanane
 * © 2023
 */
func main() {
	ctx, cancelCtxFn := context.WithCancel(context.Background())

	cfg := app.SetupEnvConfig()

	fmt.Printf("Running master")

	// Intercepting shutdown signals.
	go waitForSignal(ctx, cancelCtxFn)

	appCtx := app.NewContext(cfg)
	logExitMsg(app.Setup(ctx, appCtx))
}

func waitForSignal(ctx context.Context, cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-signals:
		log.Printf("received signal: %s, exiting gracefully", s)
		cancel()
	case <-ctx.Done():
		log.Printf("Service context done, serving remaining requests and exiting.")
	}
}

func logExitMsg(err error) {
	if err != nil {
		log.Fatalf("Service failed to setup: %s", err)
	}

	log.Print("Service exited successfully")
}
