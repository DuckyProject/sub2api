package service

import "time"

// Payment provider constants
const (
	PaymentProviderEpay     = "epay"
	PaymentProviderTokenPay = "tokenpay"
	PaymentProviderManual   = "manual"
)

// Payment kind constants
const (
	PaymentKindSubscription = "subscription"
	PaymentKindBalance      = "balance"
)

// Payment product status
const (
	PaymentProductStatusActive   = "active"
	PaymentProductStatusInactive = "inactive"
)

// Payment order status
const (
	PaymentOrderStatusCreated   = "created"
	PaymentOrderStatusPaid      = "paid"
	PaymentOrderStatusFulfilled = "fulfilled"
	PaymentOrderStatusCancelled = "cancelled"
	PaymentOrderStatusExpired   = "expired"
	PaymentOrderStatusFailed    = "failed"
)

type PaymentProduct struct {
	ID int64

	Kind          string
	Name          string
	DescriptionMD string
	Status        string
	SortOrder     int

	Currency   string
	PriceCents int64

	GroupID       *int64
	ValidityDays  *int
	CreditBalance *float64

	AllowCustomAmount    bool
	MinAmountCents       *int64
	MaxAmountCents       *int64
	SuggestedAmountsCents []int64
	ExchangeRate         *float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentOrder struct {
	ID int64

	OrderNo string
	UserID  int64
	Kind    string
	ProductID *int64

	Status   string
	Provider string

	Currency    string
	AmountCents int64

	ClientRequestID  *string
	ProviderTradeNo  *string
	PayURL           *string
	ExpiresAt        *time.Time
	PaidAt           *time.Time
	FulfilledAt      *time.Time

	GrantGroupID       *int64
	GrantValidityDays  *int
	GrantCreditBalance *float64

	Notes     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentNotification struct {
	ID int64

	Provider string
	EventID  string

	OrderNo         *string
	ProviderTradeNo *string
	AmountCents     *int64
	Currency        *string

	Verified     bool
	Processed    bool
	ProcessError *string

	RawBody     string
	ReceivedAt  time.Time
}

type EntitlementEvent struct {
	ID int64

	UserID  int64
	Kind    string
	Source  string

	GroupID         *int64
	ValidityDays    *int
	BalanceDelta    *float64
	ConcurrencyDelta *int

	OrderID      *int64
	RedeemCodeID *int64
	ActorUserID  *int64

	Note      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
