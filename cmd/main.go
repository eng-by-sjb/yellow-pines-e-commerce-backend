package main

import (
	"fmt"
	"log"

	"github.com/dev-by-sjb/e-commerce-vsa/cmd/api"
	"github.com/dev-by-sjb/e-commerce-vsa/internal/config"
)

var (
	srvPort = config.Env.ServerAddr
)

func main() {
	srv := api.NewServer(srvPort, nil)

	if err := srv.Start(); err != nil {
		log.Fatal(fmt.Errorf("failed to start server: %w", err))
	}
}
