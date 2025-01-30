package main

import (
	"context"
	"fmt"
	"goproxy/internal/app"
	"goproxy/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		log.Fatal(cfgErr)
	}

	application := app.NewApp(cfg)

	cxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("Received: %v. Shutting down...", sig)
		cancel()
	}()

	appErr := application.Run(cxt)
	if appErr != nil {
		log.Fatal(appErr)
	}
}
