package main

import (
	"diplom_part1/internal/config"
	"diplom_part1/internal/handlers"
	"diplom_part1/internal/store"
	"diplom_part1/internal/workers"
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

	// data base
	if err := store.DBInit(&cfg); err != nil {
		log.Fatal(err)
	}
	defer cfg.ConnectDB.Close()

	// router
	router := handlers.NewRouter(cfg)

	// server
	server := createServer(cfg, router)

	// workers
	workers.StartWorkers(cfg)
	defer workers.CloseWorkers(cfg)

	// listen
	log.Fatal(server.ListenAndServe())

}
