package main

import (
	"fmt"
	"log"

	"github.com/dev-by-sjb/yellow-pines-e-commerce-backend/cmd/server"
	"github.com/dev-by-sjb/yellow-pines-e-commerce-backend/internal/config"
)

var (
	srvAddr = config.Env.ServerAddr
)

func main() {
	log.SetFlags(log.Ldate | log.Lshortfile)
	srv := server.NewServer(srvAddr, nil)

	if err := srv.Start(); err != nil {
		log.Fatal(fmt.Errorf("failed to start server: %w", err))
	}
}
