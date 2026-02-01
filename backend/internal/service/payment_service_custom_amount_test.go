//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type paymentSettingRepoStub struct {
	values map[string]string
}

func (s *paymentSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	_ = ctx
	_ = key
	return nil, ErrSettingNotFound
}

func (s *paymentSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	_ = ctx
	if s.values == nil {
		return "", ErrSettingNotFound
	}
	v, ok := s.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return v, nil
}

func (s *paymentSettingRepoStub) Set(ctx context.Context, key, value string) error {
	_ = ctx
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *paymentSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	_ = ctx
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := s.values[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

func (s *paymentSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	_ = ctx
	if s.values == nil {
		s.values = map[string]string{}
	}
	for k, v := range settings {
		s.values[k] = v
	}
	return nil
}

func (s *paymentSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	_ = ctx
	out := make(map[string]string, len(s.values))
	for k, v := range s.values {
		out[k] = v
	}
	return out, nil
}

func (s *paymentSettingRepoStub) Delete(ctx context.Context, key string) error {
	_ = ctx
	delete(s.values, key)
	return nil
}

type paymentProductRepoStub struct {
	product *PaymentProduct
}

func (s *paymentProductRepoStub) Create(ctx context.Context, p *PaymentProduct) error {
	_ = ctx
	_ = p
	return errors.New("not implemented")
}

func (s *paymentProductRepoStub) Update(ctx context.Context, p *PaymentProduct) error {
	_ = ctx
	_ = p
	return errors.New("not implemented")
}

func (s *paymentProductRepoStub) Delete(ctx context.Context, id int64) error {
	_ = ctx
	_ = id
	return errors.New("not implemented")
}

func (s *paymentProductRepoStub) GetByID(ctx context.Context, id int64) (*PaymentProduct, error) {
	_ = ctx
	if s.product == nil || s.product.ID != id {
		return nil, &dbent.NotFoundError{}
	}
	clone := *s.product
	return &clone, nil
}

func (s *paymentProductRepoStub) ListActiveByKind(ctx context.Context, kind string) ([]PaymentProduct, error) {
	_ = ctx
	_ = kind
	return nil, errors.New("not implemented")
}

func (s *paymentProductRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, kind, status, search string) ([]PaymentProduct, *pagination.PaginationResult, error) {
	_ = ctx
	_ = params
	_ = kind
	_ = status
	_ = search
	return nil, nil, errors.New("not implemented")
}

type paymentOrderRepoStub struct {
	created []*PaymentOrder
}

func (s *paymentOrderRepoStub) Create(ctx context.Context, o *PaymentOrder) error {
	_ = ctx
	if o == nil {
		return nil
	}
	clone := *o
	clone.ID = int64(len(s.created) + 1)
	s.created = append(s.created, &clone)
	o.ID = clone.ID
	o.CreatedAt = time.Now()
	o.UpdatedAt = o.CreatedAt
	return nil
}

func (s *paymentOrderRepoStub) GetByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error) {
	_ = ctx
	_ = orderNo
	return nil, &dbent.NotFoundError{}
}

func (s *paymentOrderRepoStub) GetByProviderTradeNo(ctx context.Context, provider, tradeNo string) (*PaymentOrder, error) {
	_ = ctx
	_ = provider
	_ = tradeNo
	return nil, &dbent.NotFoundError{}
}

func (s *paymentOrderRepoStub) GetByUserAndClientRequestID(ctx context.Context, userID int64, clientRequestID string) (*PaymentOrder, error) {
	_ = ctx
	_ = userID
	_ = clientRequestID
	return nil, &dbent.NotFoundError{}
}

func (s *paymentOrderRepoStub) LockByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error) {
	_ = ctx
	_ = orderNo
	return nil, &dbent.NotFoundError{}
}

func (s *paymentOrderRepoStub) Update(ctx context.Context, o *PaymentOrder) error {
	_ = ctx
	_ = o
	return errors.New("not implemented")
}

func (s *paymentOrderRepoStub) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	_ = ctx
	_ = userID
	_ = params
	_ = status
	return nil, nil, errors.New("not implemented")
}

func (s *paymentOrderRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, userID *int64, kind, status, provider, search string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	_ = ctx
	_ = params
	_ = userID
	_ = kind
	_ = status
	_ = provider
	_ = search
	return nil, nil, errors.New("not implemented")
}

type paymentProviderStub struct{}

func (p *paymentProviderStub) Name() string { return "stub" }

func (p *paymentProviderStub) CreatePayment(ctx context.Context, order *PaymentOrder) (string, string, error) {
	_ = ctx
	if order == nil {
		return "", "", errors.New("nil order")
	}
	return "https://pay.example.com", "T123", nil
}

func (p *paymentProviderStub) VerifyAndParseNotify(ctx context.Context, rawBody []byte, headers map[string][]string, query map[string][]string) (*ProviderNotifyEvent, error) {
	_ = ctx
	_ = rawBody
	_ = headers
	_ = query
	return nil, errors.New("not implemented")
}

func (p *paymentProviderStub) SuccessResponse() (int, string, string) {
	return 200, "success", "text/plain"
}

func TestPaymentService_CreateOrder_BalanceCustomAmount_UsesRequestedAmount(t *testing.T) {
	settingSvc := NewSettingService(&paymentSettingRepoStub{
		values: map[string]string{
			SettingKeyPaymentEnabled:             "true",
			SettingKeyPaymentBalanceExchangeRate: "2",
		},
	}, nil)

	min := int64(100)  // 1.00
	max := int64(2000) // 20.00
	exchangeRate := 2.0
	product := &PaymentProduct{
		ID:                    10,
		Kind:                  PaymentKindBalance,
		Name:                  "Topup",
		Status:                PaymentProductStatusActive,
		Currency:              "CNY",
		PriceCents:            0,
		AllowCustomAmount:     true,
		MinAmountCents:        &min,
		MaxAmountCents:        &max,
		SuggestedAmountsCents: []int64{500, 1000},
		ExchangeRate:          &exchangeRate,
	}

	productRepo := &paymentProductRepoStub{product: product}
	orderRepo := &paymentOrderRepoStub{}

	svc := NewPaymentService(
		nil,
		settingSvc,
		productRepo,
		orderRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		map[string]PaymentProvider{"stub": &paymentProviderStub{}},
	)

	rawAmount := "5.00"
	order, err := svc.CreateOrder(context.Background(), 1, 10, "stub", nil, &rawAmount, nil)
	require.NoError(t, err)
	require.Equal(t, int64(500), order.AmountCents)
	require.NotNil(t, order.GrantCreditBalance)
	require.InDelta(t, 10.0, *order.GrantCreditBalance, 1e-9)
}

func TestPaymentService_CreateOrder_BalanceCustomAmount_RequiresAmount(t *testing.T) {
	settingSvc := NewSettingService(&paymentSettingRepoStub{
		values: map[string]string{
			SettingKeyPaymentEnabled: "true",
		},
	}, nil)

	min := int64(100)
	max := int64(2000)
	product := &PaymentProduct{
		ID:                10,
		Kind:              PaymentKindBalance,
		Name:              "Topup",
		Status:            PaymentProductStatusActive,
		Currency:          "CNY",
		AllowCustomAmount: true,
		MinAmountCents:    &min,
		MaxAmountCents:    &max,
	}

	svc := NewPaymentService(
		nil,
		settingSvc,
		&paymentProductRepoStub{product: product},
		&paymentOrderRepoStub{},
		nil, nil, nil, nil, nil, nil,
		map[string]PaymentProvider{"stub": &paymentProviderStub{}},
	)

	_, err := svc.CreateOrder(context.Background(), 1, 10, "stub", nil, nil, nil)
	require.ErrorIs(t, err, ErrPaymentOrderAmountRequired)
}

func TestPaymentService_CreateOrder_BalanceCustomAmount_OutOfRange(t *testing.T) {
	settingSvc := NewSettingService(&paymentSettingRepoStub{
		values: map[string]string{
			SettingKeyPaymentEnabled: "true",
		},
	}, nil)

	min := int64(1000) // 10.00
	max := int64(2000) // 20.00
	product := &PaymentProduct{
		ID:                10,
		Kind:              PaymentKindBalance,
		Name:              "Topup",
		Status:            PaymentProductStatusActive,
		Currency:          "CNY",
		AllowCustomAmount: true,
		MinAmountCents:    &min,
		MaxAmountCents:    &max,
	}

	svc := NewPaymentService(
		nil,
		settingSvc,
		&paymentProductRepoStub{product: product},
		&paymentOrderRepoStub{},
		nil, nil, nil, nil, nil, nil,
		map[string]PaymentProvider{"stub": &paymentProviderStub{}},
	)

	rawAmount := "5.00"
	_, err := svc.CreateOrder(context.Background(), 1, 10, "stub", nil, &rawAmount, nil)
	require.ErrorIs(t, err, ErrPaymentOrderAmountOutOfRange)
}

func TestPaymentService_CreateOrder_BalanceFixedAmount_DisallowsCustomAmount(t *testing.T) {
	settingSvc := NewSettingService(&paymentSettingRepoStub{
		values: map[string]string{
			SettingKeyPaymentEnabled: "true",
		},
	}, nil)

	product := &PaymentProduct{
		ID:                10,
		Kind:              PaymentKindBalance,
		Name:              "Topup Fixed",
		Status:            PaymentProductStatusActive,
		Currency:          "CNY",
		PriceCents:        9900,
		AllowCustomAmount: false,
	}

	svc := NewPaymentService(
		nil,
		settingSvc,
		&paymentProductRepoStub{product: product},
		&paymentOrderRepoStub{},
		nil, nil, nil, nil, nil, nil,
		map[string]PaymentProvider{"stub": &paymentProviderStub{}},
	)

	rawAmount := "5.00"
	_, err := svc.CreateOrder(context.Background(), 1, 10, "stub", nil, &rawAmount, nil)
	require.ErrorIs(t, err, ErrPaymentOrderAmountNotAllowed)
}
