package main

import (
	"diplom_part1/internal/config"
	"diplom_part1/internal/handlers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func createServer(cfg config.Config, router *chi.Mux) *http.Server {
	server := http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	return &server
}

func main() {

	// config
	cfg := config.New()

	cfgHndl := handlers.NewConfig(cfg)

	// router
	router := cfgHndl.NewRouter()

	// server
	server := createServer(cfg, router)

	// workers
	cfgHndl.StartWorkers(cfg)
	defer cfgHndl.CloseWorkers()
	defer cfgHndl.CloseDB()

	// listen
	log.Fatal(server.ListenAndServe())

}
