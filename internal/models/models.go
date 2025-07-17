package models

import "time"

// Transaction представляет модель транзакции
// @Description Детальная информация о транзакции между кошельками
type Transaction struct {
	SenderAddress   string    `json:"sender_address" validate:"required"`
	RecieverAddress string    `json:"receiver_address" validate:"required"`
	Amount          float64   `json:"amount" validate:"required"`
	Date            time.Time `json:"date" example:"2023-05-15T10:00:00Z" format:"date-time"`
}
