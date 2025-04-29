package main

import (
	"context"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
	"log"

	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/ethclient"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/blockreward"
	"github.com/powerslider/ethereum-validator-api/pkg/config"
	"github.com/powerslider/ethereum-validator-api/pkg/handlers"
	"github.com/powerslider/ethereum-validator-api/pkg/server"
	"github.com/powerslider/ethereum-validator-api/pkg/syncduties"
)

// @title Ethereum Validator API
// @version 1.0
// @description Provides validator block rewards and sync duties information.
// @termsOfService http://swagger.io/terms/

// @contact.name Tsvetan Dimitrov
// @contact.email tsvetan.dimitrov23@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1
func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatalf("[Fatal] %v", err)
	}
}

func run(ctx context.Context) error {
	err := godotenv.Load()
	if err != nil {
		return pkgerrors.Wrap(err, "load .env file")
	}

	// Load config.
	cfg, err := config.Load()
	if err != nil {
		return pkgerrors.Wrap(err, "load config")
	}

	ethClient, err := ethclient.Dial(cfg.RPCEndpoint)
	if err != nil {
		return pkgerrors.Wrap(err, "connect to execution client")
	}
	defer ethClient.Close()

	// Initialize services, router and server.
	beaconSvc := beacon.NewService(cfg.RPCEndpoint)
	blockRewardSvc := blockreward.NewService(ethClient, beaconSvc)
	syncDutySvc := syncduties.NewService(beaconSvc)

	r := handlers.SetupRouter(blockRewardSvc, syncDutySvc)
	srv := server.NewServer(cfg, r)

	// Run server.
	if err = srv.Run(ctx); err != nil {
		return pkgerrors.Wrap(err, "server exited")
	}

	log.Println("[Exit] Server shut down gracefully")

	return nil
}
