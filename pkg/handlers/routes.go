package handlers

import (
	"github.com/gorilla/mux"
	"github.com/powerslider/ethereum-validator-api/pkg/blockreward"
	"github.com/powerslider/ethereum-validator-api/pkg/syncduties"
	httpswagger "github.com/swaggo/http-swagger"

	_ "github.com/powerslider/ethereum-validator-api/docs" // generated docs
)

// SetupRouter configures all routes and returns a mux.Router.
func SetupRouter(blockRewardSvc *blockreward.Service, syncDutySvc *syncduties.Service) *mux.Router {
	r := mux.NewRouter()

	// API v1 subrouter
	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	apiV1.HandleFunc("/blockreward/{slot}", GetBlockRewardHandler(blockRewardSvc)).Methods("GET")
	apiV1.HandleFunc("/syncduties/{slot}", GetSyncDutiesHandler(syncDutySvc)).Methods("GET")
	// Swagger endpoint
	r.PathPrefix("/swagger/").Handler(httpswagger.WrapHandler)

	return r
}
