package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-seidon/chariot/internal/app"
	"github.com/go-seidon/chariot/internal/restapp"
)

func main() {
	config, err := app.NewDefaultConfig()
	if err != nil {
		panic(err)
	}

	app, err := restapp.NewRestApp(
		restapp.WithConfig(config),
	)
	if err != nil {
		panic(err)
	}

	go func() {
		err := app.Run(context.Background())
		if err != nil {
			log.Fatalf("failed running app %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = app.Stop(ctx)
	if err != nil {
		log.Fatalf("failed stopping app %v", err)
	}
}
