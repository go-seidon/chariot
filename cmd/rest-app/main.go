package main

import (
	"context"

	"github.com/go-seidon/chariot/internal/app"
	rest_app "github.com/go-seidon/chariot/internal/rest-app"
)

func main() {
	config, err := app.NewDefaultConfig()
	if err != nil {
		panic(err)
	}

	app, err := rest_app.NewRestApp(
		rest_app.WithConfig(config),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	err = app.Run(ctx)
	if err != nil {
		panic(err)
	}
}
