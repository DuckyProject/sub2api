package dto

import "time"

type PaymentProduct struct {
	ID            int64  `json:"id"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	DescriptionMD string `json:"description_md"`
	Status        string `json:"status"`
	SortOrder     int    `json:"sort_order"`

	Currency   string `json:"currency"`
	PriceCents int64  `json:"price_cents"`

	GroupID       *int64   `json:"group_id"`
	ValidityDays  *int     `json:"validity_days"`
	CreditBalance *float64 `json:"credit_balance"`

	AllowCustomAmount     bool     `json:"allow_custom_amount"`
	MinAmountCents        *int64   `json:"min_amount_cents"`
	MaxAmountCents        *int64   `json:"max_amount_cents"`
	SuggestedAmountsCents []int64  `json:"suggested_amounts_cents"`
	ExchangeRate          *float64 `json:"exchange_rate"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentOrder struct {
	ID              int64      `json:"id"`
	OrderNo         string     `json:"order_no"`
	UserID          int64      `json:"user_id"`
	Kind            string     `json:"kind"`
	ProductID       *int64     `json:"product_id"`
	Status          string     `json:"status"`
	Provider        string     `json:"provider"`
	Currency        string     `json:"currency"`
	AmountCents     int64      `json:"amount_cents"`
	ClientRequestID *string    `json:"client_request_id"`
	ProviderTradeNo *string    `json:"provider_trade_no"`
	PayURL          *string    `json:"pay_url"`
	ExpiresAt       *time.Time `json:"expires_at"`
	PaidAt          *time.Time `json:"paid_at"`
	FulfilledAt     *time.Time `json:"fulfilled_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type PaymentNotification struct {
	ID              int64     `json:"id"`
	Provider        string    `json:"provider"`
	EventID         string    `json:"event_id"`
	OrderNo         *string   `json:"order_no"`
	ProviderTradeNo *string   `json:"provider_trade_no"`
	AmountCents     *int64    `json:"amount_cents"`
	Currency        *string   `json:"currency"`
	Verified        bool      `json:"verified"`
	Processed       bool      `json:"processed"`
	ProcessError    *string   `json:"process_error"`
	RawBody         string    `json:"raw_body"`
	ReceivedAt      time.Time `json:"received_at"`
}

type CreatePaymentOrderRequest struct {
	ProductID       int64   `json:"product_id"`
	Provider        string  `json:"provider"`
	ClientRequestID *string `json:"client_request_id"`
	// Amount / AmountCents 用于余额充值自定义金额下单（allow_custom_amount=true）。
	// - amount: 例如 "12.34"
	// - amount_cents: 例如 1234（优先生效）
	Amount      *string `json:"amount"`
	AmountCents *int64  `json:"amount_cents"`
}

type CreatePaymentOrderResponse struct {
	Order PaymentOrder `json:"order"`
}
