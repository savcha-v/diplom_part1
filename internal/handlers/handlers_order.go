package handlers

import (
	"diplom_part1/internal/cookie"
	"diplom_part1/internal/encryption"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (config *ConfigHndl) getBalance(w http.ResponseWriter, r *http.Request) {

	type out struct {
		Balanse  float32 `json:"current"`
		SumSpent float32 `json:"withdrawn"`
	}

	userID := cookie.GetCookie(r, config.Key, "userID")

	balanse, spent, err := config.DB.GetBalanseSpent(r.Context(), userID)
	if err != nil {
		http.Error(w, "data base error", http.StatusInternalServerError)
		return
	}

	valueOut := out{}
	valueOut.Balanse = balanse
	valueOut.SumSpent = spent

	result, err := json.Marshal(valueOut)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
	fmt.Fprint(w)
}

func (config *ConfigHndl) postOrder(w http.ResponseWriter, r *http.Request) {

	// 200 — номер заказа уже был загружен этим пользователем;
	// 202 — новый номер заказа принят в обработку;
	// 400 — неверный формат запроса;
	// 401 — пользователь не аутентифицирован;
	// 409 — номер заказа уже был загружен другим пользователем;
	// 422 — неверный формат номера заказа;
	// 500 — внутренняя ошибка сервера.

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := string(body)
	if order == "" {
		http.Error(w, "order not found", http.StatusBadRequest)
		return
	}

	if !encryption.CheckOrder(order) {
		http.Error(w, "invalid format order", http.StatusUnprocessableEntity)
		return
	}

	userID := cookie.GetCookie(r, config.Key, "userID")
	httpStatus, order := config.DB.AddOrder(r.Context(), order, userID, config.ChanOrdersProc)
	if order != "" {
		config.AddOrderToChannelProc(order)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	w.Write([]byte(""))
	fmt.Fprint(w)
}

func (config *ConfigHndl) getOrders(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	userID := cookie.GetCookie(r, config.Key, "userID")

	valueOut, err := config.DB.GetAccum(r.Context(), userID)
	if err != nil {
		http.Error(w, "getOrders/ data base error", http.StatusInternalServerError)
		return
	}

	if len(valueOut) == 0 {
		http.Error(w, "getOrders/ no orders", http.StatusNoContent)
		return
	}

	result, err := json.Marshal(valueOut)
	if err != nil {
		http.Error(w, "getOrders/ marshal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
	fmt.Fprint(w)
}

func (config *ConfigHndl) postWithdraw(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type in struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}

	valueIn := in{}

	if err := json.Unmarshal(body, &valueIn); err != nil || valueIn.Order == "" || valueIn.Sum == 0 {
		http.Error(w, "unmarshal error", http.StatusBadRequest)
		return
	}

	if !encryption.CheckOrder(valueIn.Order) {
		http.Error(w, "invalid format order", http.StatusUnprocessableEntity)
		return
	}

	userID := cookie.GetCookie(r, config.Key, "userID")
	httpStatus := config.DB.WriteWithdraw(r.Context(), valueIn.Order, valueIn.Sum, userID)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	w.Write([]byte(""))
	fmt.Fprint(w)
}

func (config *ConfigHndl) getWithdrawals(w http.ResponseWriter, r *http.Request) {

	userID := cookie.GetCookie(r, config.Key, "userID")

	valueOut, err := config.DB.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "data base error", http.StatusInternalServerError)
		return
	}

	if len(valueOut) == 0 {
		http.Error(w, "no orders", http.StatusNoContent)
		return
	}

	result, err := json.Marshal(valueOut)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
	fmt.Fprint(w)
}
