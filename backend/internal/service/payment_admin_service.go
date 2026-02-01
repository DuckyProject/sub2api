package service

import (
	"context"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPaymentAdminProductNotFound = infraerrors.NotFound("PAYMENT_PRODUCT_NOT_FOUND", "payment product not found")
)

type PaymentAdminService struct {
	productRepo PaymentProductRepository
	orderRepo   PaymentOrderRepository
	notifyRepo  PaymentNotificationRepository
	groupRepo   GroupRepository
}

func NewPaymentAdminService(
	productRepo PaymentProductRepository,
	orderRepo PaymentOrderRepository,
	notifyRepo PaymentNotificationRepository,
	groupRepo GroupRepository,
) *PaymentAdminService {
	return &PaymentAdminService{
		productRepo: productRepo,
		orderRepo:   orderRepo,
		notifyRepo:  notifyRepo,
		groupRepo:   groupRepo,
	}
}

type CreatePaymentProductInput struct {
	Kind          string
	Name          string
	DescriptionMD string
	Status        string
	SortOrder     int

	Currency   string
	PriceCents int64

	GroupID      *int64
	ValidityDays *int

	CreditBalance *float64

	AllowCustomAmount     bool
	MinAmountCents        *int64
	MaxAmountCents        *int64
	SuggestedAmountsCents []int64
	ExchangeRate          *float64
}

type UpdatePaymentProductInput struct {
	Kind          *string
	Name          *string
	DescriptionMD *string
	Status        *string
	SortOrder     *int

	Currency   *string
	PriceCents *int64

	GroupID      *int64
	ValidityDays *int

	CreditBalance *float64

	AllowCustomAmount     *bool
	MinAmountCents        *int64
	MaxAmountCents        *int64
	SuggestedAmountsCents *[]int64
	ExchangeRate          *float64
}

func (s *PaymentAdminService) ListProducts(ctx context.Context, params pagination.PaginationParams, kind, status, search string) ([]PaymentProduct, *pagination.PaginationResult, error) {
	return s.productRepo.ListWithFilters(ctx, params, strings.TrimSpace(kind), strings.TrimSpace(status), strings.TrimSpace(search))
}

func (s *PaymentAdminService) GetProduct(ctx context.Context, id int64) (*PaymentProduct, error) {
	p, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPaymentAdminProductNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PaymentAdminService) DeleteProduct(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrPaymentProductInvalid
	}
	_, err := s.GetProduct(ctx, id)
	if err != nil {
		return err
	}
	if err := s.productRepo.Delete(ctx, id); err != nil {
		if dbent.IsNotFound(err) {
			return ErrPaymentAdminProductNotFound
		}
		return fmt.Errorf("delete payment product: %w", err)
	}
	return nil
}

func (s *PaymentAdminService) ListOrders(ctx context.Context, params pagination.PaginationParams, userID *int64, kind, status, provider, search string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	if s.orderRepo == nil {
		return nil, nil, fmt.Errorf("nil payment order repository")
	}
	return s.orderRepo.ListWithFilters(ctx, params, userID, strings.TrimSpace(kind), strings.TrimSpace(status), strings.TrimSpace(provider), strings.TrimSpace(search))
}

func (s *PaymentAdminService) GetOrder(ctx context.Context, orderNo string) (*PaymentOrder, error) {
	if s.orderRepo == nil {
		return nil, fmt.Errorf("nil payment order repository")
	}
	o, err := s.orderRepo.GetByOrderNo(ctx, strings.TrimSpace(orderNo))
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPaymentOrderNotFound
		}
		return nil, err
	}
	return o, nil
}

func (s *PaymentAdminService) ListNotifications(ctx context.Context, params pagination.PaginationParams, provider, orderNo, search string) ([]PaymentNotification, *pagination.PaginationResult, error) {
	if s.notifyRepo == nil {
		return nil, nil, fmt.Errorf("nil payment notification repository")
	}
	return s.notifyRepo.ListWithFilters(ctx, params, strings.TrimSpace(provider), strings.TrimSpace(orderNo), strings.TrimSpace(search))
}

func (s *PaymentAdminService) CreateProduct(ctx context.Context, input *CreatePaymentProductInput) (*PaymentProduct, error) {
	if input == nil {
		return nil, ErrPaymentProductInvalid
	}

	p := &PaymentProduct{
		Kind:                  strings.ToLower(strings.TrimSpace(input.Kind)),
		Name:                  strings.TrimSpace(input.Name),
		DescriptionMD:         strings.TrimSpace(input.DescriptionMD),
		Status:                strings.ToLower(strings.TrimSpace(input.Status)),
		SortOrder:             input.SortOrder,
		Currency:              strings.ToUpper(strings.TrimSpace(input.Currency)),
		PriceCents:            input.PriceCents,
		GroupID:               input.GroupID,
		ValidityDays:          input.ValidityDays,
		CreditBalance:         input.CreditBalance,
		AllowCustomAmount:     input.AllowCustomAmount,
		MinAmountCents:        input.MinAmountCents,
		MaxAmountCents:        input.MaxAmountCents,
		SuggestedAmountsCents: input.SuggestedAmountsCents,
		ExchangeRate:          input.ExchangeRate,
	}

	// Normalize pointer fields (Create 请求与 Update 行为保持一致)。
	if p.GroupID != nil && *p.GroupID <= 0 {
		p.GroupID = nil
	}
	if p.ValidityDays != nil && *p.ValidityDays <= 0 {
		p.ValidityDays = nil
	}
	if p.CreditBalance != nil && *p.CreditBalance <= 0 {
		p.CreditBalance = nil
	}
	if p.MinAmountCents != nil && *p.MinAmountCents <= 0 {
		p.MinAmountCents = nil
	}
	if p.MaxAmountCents != nil && *p.MaxAmountCents <= 0 {
		p.MaxAmountCents = nil
	}
	if p.ExchangeRate != nil && *p.ExchangeRate <= 0 {
		p.ExchangeRate = nil
	}

	s.applyProductDefaults(p)
	if err := s.validateProduct(ctx, p); err != nil {
		return nil, err
	}

	if err := s.productRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create payment product: %w", err)
	}
	return p, nil
}

func (s *PaymentAdminService) UpdateProduct(ctx context.Context, id int64, input *UpdatePaymentProductInput) (*PaymentProduct, error) {
	if input == nil {
		return nil, ErrPaymentProductInvalid
	}
	p, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPaymentAdminProductNotFound
		}
		return nil, err
	}

	if input.Kind != nil {
		p.Kind = strings.ToLower(strings.TrimSpace(*input.Kind))
	}
	if input.Name != nil {
		p.Name = strings.TrimSpace(*input.Name)
	}
	if input.DescriptionMD != nil {
		p.DescriptionMD = strings.TrimSpace(*input.DescriptionMD)
	}
	if input.Status != nil {
		p.Status = strings.ToLower(strings.TrimSpace(*input.Status))
	}
	if input.SortOrder != nil {
		p.SortOrder = *input.SortOrder
	}
	if input.Currency != nil {
		p.Currency = strings.ToUpper(strings.TrimSpace(*input.Currency))
	}
	if input.PriceCents != nil {
		p.PriceCents = *input.PriceCents
	}
	if input.GroupID != nil {
		if *input.GroupID <= 0 {
			p.GroupID = nil
		} else {
			v := *input.GroupID
			p.GroupID = &v
		}
	}
	if input.ValidityDays != nil {
		if *input.ValidityDays <= 0 {
			p.ValidityDays = nil
		} else {
			v := *input.ValidityDays
			p.ValidityDays = &v
		}
	}
	if input.CreditBalance != nil {
		if *input.CreditBalance <= 0 {
			p.CreditBalance = nil
		} else {
			v := *input.CreditBalance
			p.CreditBalance = &v
		}
	}
	if input.AllowCustomAmount != nil {
		p.AllowCustomAmount = *input.AllowCustomAmount
	}
	if input.MinAmountCents != nil {
		if *input.MinAmountCents <= 0 {
			p.MinAmountCents = nil
		} else {
			v := *input.MinAmountCents
			p.MinAmountCents = &v
		}
	}
	if input.MaxAmountCents != nil {
		if *input.MaxAmountCents <= 0 {
			p.MaxAmountCents = nil
		} else {
			v := *input.MaxAmountCents
			p.MaxAmountCents = &v
		}
	}
	if input.SuggestedAmountsCents != nil {
		p.SuggestedAmountsCents = *input.SuggestedAmountsCents
	}
	if input.ExchangeRate != nil {
		if *input.ExchangeRate <= 0 {
			p.ExchangeRate = nil
		} else {
			v := *input.ExchangeRate
			p.ExchangeRate = &v
		}
	}

	s.applyProductDefaults(p)
	if err := s.validateProduct(ctx, p); err != nil {
		return nil, err
	}

	if err := s.productRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update payment product: %w", err)
	}
	return p, nil
}

func (s *PaymentAdminService) applyProductDefaults(p *PaymentProduct) {
	if p == nil {
		return
	}
	if strings.TrimSpace(p.Status) == "" {
		p.Status = PaymentProductStatusInactive
	}
	if strings.TrimSpace(p.Currency) == "" {
		p.Currency = "CNY"
	}
	if p.SuggestedAmountsCents == nil {
		p.SuggestedAmountsCents = []int64{}
	}
}

func (s *PaymentAdminService) validateProduct(ctx context.Context, p *PaymentProduct) error {
	if p == nil {
		return ErrPaymentProductInvalid
	}
	if p.Kind != PaymentKindSubscription && p.Kind != PaymentKindBalance {
		return ErrPaymentProductInvalid
	}
	if p.Name == "" {
		return ErrPaymentProductInvalid
	}
	if p.Status != PaymentProductStatusActive && p.Status != PaymentProductStatusInactive {
		return ErrPaymentProductInvalid
	}
	if strings.TrimSpace(p.Currency) == "" {
		return ErrPaymentProductInvalid
	}
	if p.PriceCents < 0 {
		return ErrPaymentProductInvalid
	}

	switch p.Kind {
	case PaymentKindSubscription:
		if p.GroupID == nil || *p.GroupID <= 0 {
			return ErrPaymentProductInvalid
		}
		if p.ValidityDays == nil || *p.ValidityDays <= 0 {
			return ErrPaymentProductInvalid
		}

		if s.groupRepo != nil {
			group, err := s.groupRepo.GetByID(ctx, *p.GroupID)
			if err != nil {
				return err
			}
			if group == nil || !group.IsSubscriptionType() {
				return ErrGroupNotSubscriptionType
			}
		}

		// Subscription products should not carry balance-specific configs.
		p.CreditBalance = nil
		p.AllowCustomAmount = false
		p.MinAmountCents = nil
		p.MaxAmountCents = nil
		p.SuggestedAmountsCents = []int64{}
		p.ExchangeRate = nil

	case PaymentKindBalance:
		// Balance products should not carry subscription-specific configs.
		p.GroupID = nil
		p.ValidityDays = nil

		if p.AllowCustomAmount {
			if p.MinAmountCents == nil || *p.MinAmountCents <= 0 {
				return ErrPaymentProductInvalid
			}
			if p.MaxAmountCents == nil || *p.MaxAmountCents <= 0 {
				return ErrPaymentProductInvalid
			}
			if *p.MinAmountCents > *p.MaxAmountCents {
				return ErrPaymentProductInvalid
			}
			for _, v := range p.SuggestedAmountsCents {
				if v < *p.MinAmountCents || v > *p.MaxAmountCents {
					return ErrPaymentProductInvalid
				}
			}
		} else {
			if p.PriceCents <= 0 {
				return ErrPaymentProductInvalid
			}
		}
	}

	return nil
}
