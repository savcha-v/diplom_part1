package handlers

import (
	"diplom_part1/internal/config"
	"diplom_part1/internal/store"
	"diplom_part1/internal/workers"
	"log"

	"github.com/go-chi/chi/v5"
)

type ConfigHndl struct {
	DB      *store.DB
	Workers *workers.ConfigWork
	Key     string
}

func NewConfig(cfg config.Config) (config *ConfigHndl) {

	// data base
	db, err := store.DBInit(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// workers
	work := workers.NewConfig()

	config = &(ConfigHndl{
		DB:      db,
		Key:     "10c57de0",
		Workers: work,
	})

	return
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
