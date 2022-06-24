package handlers

import (
	"context"
	"diplom_part1/internal/store"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type OrderData struct {
	Order  string  `json:"order"`
	Status string  `json:"status"`
	Sum    float32 `json:"accrual"`
	UserID string
}

// записать в канал заказы со статусами new, registered, processing
func WriteOrderProcessing(ctx context.Context, cfg *ConfigHndl) {
	orders, err := cfg.DB.GetOrdersProcessing(ctx)
	if err != nil {
		log.Println(err)
	}
	for _, number := range orders {
		cfg.AddOrderToChannelProc(number)
	}
}

// обработать заказы из канала
func ReadOrderProcessing(ctx context.Context, cfg *ConfigHndl, accrualAddress string) {

	for number := range cfg.ChanOrdersProc {
		orderData, err := getOrderData(ctx, *cfg.DB, accrualAddress, number)
		if err != nil {
			log.Println(err)
			cfg.AddOrderToChannelProc(number)
			return
		}

		status, err := cfg.DB.UpdateOrder(ctx, orderData.UserID, orderData.Order, orderData.Status, orderData.Sum)
		if err != nil {
			log.Println(err)
			cfg.AddOrderToChannelProc(number)
			return
		}

		// если не в конечном статусе
		if status != cfg.OrdersStatus.Processed && status != cfg.OrdersStatus.Invalid {
			go cfg.AddOrderToChannelProc(number)
		}

	}
}

func getOrderData(ctx context.Context, db store.DB, accrualAddress string, number string) (OrderData, error) {

	valueIn := OrderData{}

	addressCalc := accrualAddress + "/api/orders/" + number
	r, err := http.Get(addressCalc)
	if err != nil {
		return valueIn, errors.New("error call /api/orders/")
	}

	if r.StatusCode == http.StatusTooManyRequests {

		retryHead := r.Header.Get("Retry-After")
		if retryHead != "" {
			retry, err := strconv.Atoi(retryHead)
			if err != nil {
				return valueIn, errors.New("error conv Retry-After /api/orders/")
			}
			time.Sleep(time.Duration(retry) * time.Second)

			return valueIn, errors.New("getOrderData/ wait retry /api/orders/")
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return valueIn, errors.New("error read body /api/orders/")
	}

	defer r.Body.Close()

	if err := json.Unmarshal(body, &valueIn); err != nil {
		return valueIn, errors.New("error unmarshal /api/orders/")
	}

	if valueIn.Order == "" {
		return valueIn, errors.New("error unmarshal valueIn.Order is empty /api/orders/")
	}

	userID, err := db.GetUserID(ctx, number)

	if err != nil {
		return valueIn, errors.New("error get user ID /api/orders/")
	}

	valueIn.UserID = userID

	// valueIn.UserID = `1c2be014-8880-4e33-aa94-6a3986253b0c`
	// valueIn.Status = "PROCESSED"
	// valueIn.Sum = "200"
	// valueIn.Order = number

	return valueIn, nil
}

func (config *ConfigHndl) AddOrderToChannelProc(number string) {
	config.ChanOrdersProc <- number
}
