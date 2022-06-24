package workers

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

func (cfgWork *ConfigWork) AddOrderToChannelProc(number string) {
	cfgWork.ChanOrdersProc <- number
}

type OrderData struct {
	Order  string  `json:"order"`
	Status string  `json:"status"`
	Sum    float32 `json:"accrual"`
	UserID string
}

// записать в канал заказы со статусами new, registered, processing
func (cfgWork *ConfigWork) WriteOrderProcessing(ctx context.Context, db *store.DB) {

	orders, err := db.GetOrdersProcessing(ctx)
	if err != nil {
		log.Println(err)
	}
	for _, number := range orders {
		cfgWork.AddOrderToChannelProc(number)
	}
}

// обработать заказы из канала
func (cfgWork *ConfigWork) ReadOrderProcessing(ctx context.Context, db *store.DB, accrualAddress string) {

	for number := range cfgWork.ChanOrdersProc {
		orderData, err := getOrderData(ctx, *db, accrualAddress, number)
		if err != nil {
			log.Println(err)
			cfgWork.AddOrderToChannelProc(number)
			return
		}

		status, err := db.UpdateOrder(ctx, orderData.UserID, orderData.Order, orderData.Status, orderData.Sum)
		if err != nil {
			log.Println(err)
			cfgWork.AddOrderToChannelProc(number)
			return
		}

		// если не в конечном статусе
		if status != "PROCESSED" && status != "INVALID" {
			go cfgWork.AddOrderToChannelProc(number)
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
