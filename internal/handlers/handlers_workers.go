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
func WriteOrderProcessing(ctx context.Context, config *ConfigHndl) {
	var statusProc []string
	statusProc = append(statusProc, config.OrdersStatus.New)
	statusProc = append(statusProc, config.OrdersStatus.Processing)
	statusProc = append(statusProc, config.OrdersStatus.Registered)

	orders, err := config.DB.GetOrdersProcessing(ctx, statusProc)
	if err != nil {
		log.Println(err)
	}
	for _, number := range orders {
		config.AddOrderToChannelProc(number)
	}
}

// обработать заказы из канала
func ReadOrderProcessing(ctx context.Context, config *ConfigHndl, accrualAddress string) {

	for number := range config.ChanOrdersProc {
		orderData, err := getOrderData(ctx, *config.DB, accrualAddress, number)
		if err != nil {
			log.Println(err)
			config.AddOrderToChannelProc(number)
			return
		}

		status, err := config.DB.UpdateOrder(ctx, orderData.UserID, orderData.Order, orderData.Status, orderData.Sum)
		if err != nil {
			log.Println(err)
			config.AddOrderToChannelProc(number)
			return
		}

		// если не в конечном статусе
		if status != config.OrdersStatus.Processed && status != config.OrdersStatus.Invalid {
			go config.AddOrderToChannelProc(number)
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
