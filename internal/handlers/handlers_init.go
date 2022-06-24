package handlers

import (
	"context"
	"diplom_part1/internal/config"
	"diplom_part1/internal/store"
	"log"

	"github.com/go-chi/chi/v5"
)

type ConfigHndl struct {
	DB             *store.DB
	Key            string
	OrdersStatus   OrdersStatus
	ChanOrdersProc chan string
}

type OrdersStatus struct {
	New        string // заказ загружен в систему, но не попал в обработку;
	Processing string // вознаграждение за заказ рассчитывается;
	Invalid    string // система расчёта вознаграждений отказала в расчёте;
	Processed  string // данные по заказу проверены и информация о расчёте успешно получена.
	Registered string // заказ зарегистрирован, но не начисление не рассчитано;
}

func NewConfig(cfg config.Config) (config *ConfigHndl) {

	// data base
	db, err := store.DBInit(cfg)
	if err != nil {
		log.Fatal(err)
	}

	statuses := OrdersStatus{
		New:        "NEW",
		Processing: "PROCESSING",
		Invalid:    "INVALID",
		Processed:  "PROCESSED",
		Registered: "REGISTERED",
	}

	config = &(ConfigHndl{
		DB:             db,
		Key:            "10c57de0",
		OrdersStatus:   statuses,
		ChanOrdersProc: make(chan string),
	})

	return
}

func (config *ConfigHndl) CloseDB() {
	config.DB.Connect.Close()
}

func (config *ConfigHndl) NewRouter() *chi.Mux {

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", config.userRegister) // регистрация пользователя;
		r.Post("/api/user/login", config.userLogin)       // аутентификация пользователя;
	})

	r.Group(func(r chi.Router) {
		r.Use(config.CheckAuthorized)
		r.Post("/api/user/orders", config.postOrder)                  // загрузка пользователем номера заказа для расчёта;
		r.Get("/api/user/orders", config.getOrders)                   // получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
		r.Get("/api/user/balance", config.getBalance)                 // получение текущего баланса счёта баллов лояльности пользователя;
		r.Post("/api/user/balance/withdraw", config.postWithdraw)     // запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
		r.Get("/api/user/balance/withdrawals", config.getWithdrawals) // получение информации о выводе средств с накопительного счёта пользователем.
	})
	return r
}

func (config *ConfigHndl) CloseWorkers() {
	close(config.ChanOrdersProc)
}

func (config *ConfigHndl) StartWorkers(baseCfg config.Config) {
	go WriteOrderProcessing(context.Background(), config)
	go ReadOrderProcessing(context.Background(), config, baseCfg.AccrualAddress)
}
