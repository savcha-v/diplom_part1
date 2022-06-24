package workers

import (
	"context"
	"diplom_part1/internal/config"
	"diplom_part1/internal/store"
)

type ConfigWork struct {
	ChanOrdersProc chan string
}

func NewConfig() (config *ConfigWork) {

	config = &(ConfigWork{
		ChanOrdersProc: make(chan string),
	})

	return
}

func (cfgWork *ConfigWork) Close() {
	close(cfgWork.ChanOrdersProc)
}

func (cfgWork *ConfigWork) Start(baseCfg config.Config, db *store.DB) {
	go cfgWork.WriteOrderProcessing(context.Background(), db)
	go cfgWork.ReadOrderProcessing(context.Background(), db, baseCfg.AccrualAddress)
}
