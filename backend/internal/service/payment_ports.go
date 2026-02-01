package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// PaymentProductRepository provides persistence access for payment products.
type PaymentProductRepository interface {
	Create(ctx context.Context, p *PaymentProduct) error
	Update(ctx context.Context, p *PaymentProduct) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*PaymentProduct, error)
	ListActiveByKind(ctx context.Context, kind string) ([]PaymentProduct, error)
	ListWithFilters(ctx context.Context, params pagination.PaginationParams, kind, status, search string) ([]PaymentProduct, *pagination.PaginationResult, error)
}

// PaymentOrderRepository provides persistence access for payment orders.
type PaymentOrderRepository interface {
	Create(ctx context.Context, o *PaymentOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error)
	GetByProviderTradeNo(ctx context.Context, provider, tradeNo string) (*PaymentOrder, error)
	GetByUserAndClientRequestID(ctx context.Context, userID int64, clientRequestID string) (*PaymentOrder, error)
	LockByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error)
	Update(ctx context.Context, o *PaymentOrder) error

	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error)
	ListWithFilters(ctx context.Context, params pagination.PaginationParams, userID *int64, kind, status, provider, search string) ([]PaymentOrder, *pagination.PaginationResult, error)
}

// PaymentNotificationRepository provides persistence access for callback audit and idempotency.
type PaymentNotificationRepository interface {
	CreateIfNotExists(ctx context.Context, n *PaymentNotification) (inserted bool, err error)
	MarkProcessed(ctx context.Context, provider, eventID string, processed bool, verified bool, processError *string) error
	ListWithFilters(ctx context.Context, params pagination.PaginationParams, provider, orderNo, search string) ([]PaymentNotification, *pagination.PaginationResult, error)
}

// EntitlementEventRepository provides persistence access for entitlement event audit.
type EntitlementEventRepository interface {
	Create(ctx context.Context, e *EntitlementEvent) error
}

// ProviderNotifyEvent represents a normalized callback event.
type ProviderNotifyEvent struct {
	Provider string
	EventID  string

	OrderNo         string
	ProviderTradeNo string
	AmountCents     int64
	Currency        string
	Paid            bool

	RawBody string
	Now     time.Time
}

// PaymentProvider defines provider operations.
type PaymentProvider interface {
	Name() string
	CreatePayment(ctx context.Context, order *PaymentOrder) (payURL string, providerTradeNo string, err error)
	VerifyAndParseNotify(ctx context.Context, rawBody []byte, headers map[string][]string, query map[string][]string) (*ProviderNotifyEvent, error)
	SuccessResponse() (statusCode int, body string, contentType string)
}
