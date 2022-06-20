package workers

import (
	"context"
	"diplom_part1/internal/config"
)

func CloseWorkers(cfg config.Config) {
	close(cfg.ChanOrdersProc)
}

func StartWorkers(cfg config.Config) {
	go WriteOrderProcessing(context.Background(), cfg)
	go ReadOrderProcessing(context.Background(), cfg)
}
