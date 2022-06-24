package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS" envDefault:"localhost:9090"`
	DataBase       string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
}

type OutAccum struct {
	Order  string    `json:"number"`
	Status string    `json:"status"`
	Sum    float32   `json:"accrual,omitempty"`
	Date   time.Time `json:"uploaded_at"`
}

type OutWithdrawals struct {
	Order string    `json:"order"`
	Sum   float32   `json:"accrual"`
	Date  time.Time `json:"processed_at"`
}

func New() Config {

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "")
	flag.StringVar(&cfg.DataBase, "d", cfg.DataBase, "")
	flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, "")

	flag.Parse()

	return cfg
}
