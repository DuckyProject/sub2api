package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPaymentNotEnabled          = infraerrors.Forbidden("PAYMENT_NOT_ENABLED", "payment is not enabled")
	ErrPaymentProductNotFound     = infraerrors.NotFound("PAYMENT_PRODUCT_NOT_FOUND", "payment product not found")
	ErrPaymentProductInactive     = infraerrors.Forbidden("PAYMENT_PRODUCT_INACTIVE", "payment product is inactive")
	ErrPaymentProductInvalid      = infraerrors.BadRequest("PAYMENT_PRODUCT_INVALID", "payment product is invalid")
	ErrPaymentUnsupportedProvider = infraerrors.BadRequest("PAYMENT_UNSUPPORTED_PROVIDER", "unsupported payment provider")

	ErrPaymentProviderNotFound   = infraerrors.NotFound("PAYMENT_PROVIDER_NOT_FOUND", "payment provider not found")
	ErrPaymentOrderNotFound      = infraerrors.NotFound("PAYMENT_ORDER_NOT_FOUND", "payment order not found")
	ErrPaymentOrderMismatch      = infraerrors.BadRequest("PAYMENT_ORDER_MISMATCH", "payment order mismatch")
	ErrPaymentOrderInvalidState  = infraerrors.Conflict("PAYMENT_ORDER_INVALID_STATE", "payment order state does not allow this operation")
	ErrPaymentNotifyProcessError = infraerrors.InternalServer("PAYMENT_NOTIFY_PROCESS_ERROR", "payment notify processing error")

	ErrPaymentOrderAmountRequired   = infraerrors.BadRequest("PAYMENT_ORDER_AMOUNT_REQUIRED", "amount is required for this product")
	ErrPaymentOrderAmountInvalid    = infraerrors.BadRequest("PAYMENT_ORDER_AMOUNT_INVALID", "invalid amount")
	ErrPaymentOrderAmountNotAllowed = infraerrors.BadRequest("PAYMENT_ORDER_AMOUNT_NOT_ALLOWED", "custom amount is not allowed for this product")
	ErrPaymentOrderAmountOutOfRange = infraerrors.BadRequest("PAYMENT_ORDER_AMOUNT_OUT_OF_RANGE", "amount is out of allowed range")
)

type PaymentService struct {
	entClient       *dbent.Client
	settingService  *SettingService
	productRepo     PaymentProductRepository
	orderRepo       PaymentOrderRepository
	notifyRepo      PaymentNotificationRepository
	entitlementRepo EntitlementEventRepository
	userRepo        UserRepository
	subService      *SubscriptionService
	billingCache    *BillingCacheService
	authInvalidator APIKeyAuthCacheInvalidator

	providers map[string]PaymentProvider
}

func NewPaymentService(
	entClient *dbent.Client,
	settingService *SettingService,
	productRepo PaymentProductRepository,
	orderRepo PaymentOrderRepository,
	notifyRepo PaymentNotificationRepository,
	entitlementRepo EntitlementEventRepository,
	userRepo UserRepository,
	subService *SubscriptionService,
	billingCache *BillingCacheService,
	authInvalidator APIKeyAuthCacheInvalidator,
	providers map[string]PaymentProvider,
) *PaymentService {
	pmap := make(map[string]PaymentProvider)
	for k, p := range providers {
		if p == nil {
			continue
		}
		pmap[k] = p
	}
	return &PaymentService{
		entClient:       entClient,
		settingService:  settingService,
		productRepo:     productRepo,
		orderRepo:       orderRepo,
		notifyRepo:      notifyRepo,
		entitlementRepo: entitlementRepo,
		userRepo:        userRepo,
		subService:      subService,
		billingCache:    billingCache,
		authInvalidator: authInvalidator,
		providers:       pmap,
	}
}

// ListProducts returns active products for a kind.
func (s *PaymentService) ListProducts(ctx context.Context, kind string) ([]PaymentProduct, error) {
	if !s.settingService.IsPaymentEnabled(ctx) {
		return nil, ErrPaymentNotEnabled
	}
	return s.productRepo.ListActiveByKind(ctx, kind)
}

// ListOrders returns user's payment orders.
func (s *PaymentService) ListOrders(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListByUser(ctx, userID, params, strings.TrimSpace(status))
}

// GetOrder returns a specific order if it belongs to the user.
func (s *PaymentService) GetOrder(ctx context.Context, userID int64, orderNo string) (*PaymentOrder, error) {
	o, err := s.orderRepo.GetByOrderNo(ctx, strings.TrimSpace(orderNo))
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPaymentOrderNotFound
		}
		return nil, err
	}
	if o.UserID != userID {
		return nil, infraerrors.Forbidden("PAYMENT_ORDER_FORBIDDEN", "cannot access this order")
	}
	return o, nil
}

// CreateOrder creates a payment order and returns provider pay_url.
func (s *PaymentService) CreateOrder(ctx context.Context, userID int64, productID int64, provider string, clientRequestID *string, amount *string, amountCents *int64) (*PaymentOrder, error) {
	if !s.settingService.IsPaymentEnabled(ctx) {
		return nil, ErrPaymentNotEnabled
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, ErrPaymentUnsupportedProvider
	}

	// Idempotency: same user + client_request_id returns existing order.
	if clientRequestID != nil {
		id := strings.TrimSpace(*clientRequestID)
		if id == "" {
			clientRequestID = nil
		} else {
			clientRequestID = &id
		}
	}
	if clientRequestID != nil {
		if existing, err := s.orderRepo.GetByUserAndClientRequestID(ctx, userID, *clientRequestID); err == nil && existing != nil {
			return existing, nil
		} else if err != nil && !dbent.IsNotFound(err) {
			return nil, err
		}
	}

	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, ErrPaymentProductNotFound
	}
	if strings.TrimSpace(p.Status) != PaymentProductStatusActive {
		return nil, ErrPaymentProductInactive
	}

	// Optional custom amount (mainly for balance products).
	var requestedAmountCents *int64
	if amountCents != nil {
		v := *amountCents
		if v <= 0 {
			return nil, ErrPaymentOrderAmountInvalid
		}
		requestedAmountCents = &v
	} else if amount != nil {
		raw := strings.TrimSpace(*amount)
		if raw != "" {
			v, err := parseMoneyToCents(raw)
			if err != nil || v <= 0 {
				return nil, ErrPaymentOrderAmountInvalid.WithCause(err)
			}
			requestedAmountCents = &v
		}
	}

	now := time.Now()
	orderNo := fmt.Sprintf("P%s%d", now.UTC().Format("20060102150405"), now.UnixNano()%1_000_000)
	orderCurrency := strings.TrimSpace(p.Currency)
	if orderCurrency == "" {
		orderCurrency = "CNY"
	}
	o := &PaymentOrder{
		OrderNo:         orderNo,
		UserID:          userID,
		Kind:            p.Kind,
		ProductID:       func() *int64 { v := p.ID; return &v }(),
		Status:          PaymentOrderStatusCreated,
		Provider:        provider,
		Currency:        orderCurrency,
		AmountCents:     p.PriceCents,
		ClientRequestID: clientRequestID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Snapshot grant fields at order creation time.
	switch p.Kind {
	case PaymentKindSubscription:
		if requestedAmountCents != nil {
			return nil, ErrPaymentOrderAmountNotAllowed
		}
		if p.GroupID == nil {
			return nil, ErrPaymentProductInvalid
		}
		o.GrantGroupID = func() *int64 { v := *p.GroupID; return &v }()
		validity := 30
		if p.ValidityDays != nil && *p.ValidityDays > 0 {
			validity = *p.ValidityDays
		}
		o.GrantValidityDays = func() *int { v := validity; return &v }()
	case PaymentKindBalance:
		if p.AllowCustomAmount {
			if requestedAmountCents == nil {
				md := map[string]string{}
				if p.MinAmountCents != nil {
					md["min_amount_cents"] = fmt.Sprint(*p.MinAmountCents)
				}
				if p.MaxAmountCents != nil {
					md["max_amount_cents"] = fmt.Sprint(*p.MaxAmountCents)
				}
				return nil, ErrPaymentOrderAmountRequired.WithMetadata(md)
			}
			if p.MinAmountCents != nil && *requestedAmountCents < *p.MinAmountCents {
				md := map[string]string{}
				md["min_amount_cents"] = fmt.Sprint(*p.MinAmountCents)
				if p.MaxAmountCents != nil {
					md["max_amount_cents"] = fmt.Sprint(*p.MaxAmountCents)
				}
				return nil, ErrPaymentOrderAmountOutOfRange.WithMetadata(md)
			}
			if p.MaxAmountCents != nil && *requestedAmountCents > *p.MaxAmountCents {
				md := map[string]string{}
				if p.MinAmountCents != nil {
					md["min_amount_cents"] = fmt.Sprint(*p.MinAmountCents)
				}
				md["max_amount_cents"] = fmt.Sprint(*p.MaxAmountCents)
				return nil, ErrPaymentOrderAmountOutOfRange.WithMetadata(md)
			}
			o.AmountCents = *requestedAmountCents
		} else if requestedAmountCents != nil {
			return nil, ErrPaymentOrderAmountNotAllowed
		}
		if !p.AllowCustomAmount && o.AmountCents <= 0 {
			return nil, ErrPaymentProductInvalid
		}

		exchangeRate := 1.0
		if p.ExchangeRate != nil && *p.ExchangeRate > 0 {
			exchangeRate = *p.ExchangeRate
		} else if settings, err := s.settingService.GetAllSettings(ctx); err == nil && settings != nil && settings.PaymentBalanceExchangeRate > 0 {
			exchangeRate = settings.PaymentBalanceExchangeRate
		}

		var credit float64
		if p.AllowCustomAmount {
			credit = (float64(o.AmountCents) / 100.0) * exchangeRate
		} else if p.CreditBalance != nil {
			credit = *p.CreditBalance
		} else {
			credit = (float64(o.AmountCents) / 100.0) * exchangeRate
		}
		o.GrantCreditBalance = func() *float64 { v := credit; return &v }()
	default:
		return nil, ErrPaymentProductInvalid
	}

	prov, ok := s.providers[provider]
	if !ok {
		return nil, ErrPaymentUnsupportedProvider
	}

	payURL, providerTradeNo, err := prov.CreatePayment(ctx, o)
	if err != nil {
		return nil, err
	}
	if payURL != "" {
		o.PayURL = &payURL
	}
	if providerTradeNo != "" {
		o.ProviderTradeNo = &providerTradeNo
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

// HandleNotify verifies provider callback, stores audit record, and fulfills paid orders.
//
// It returns provider-specific "success" response on successful processing (or idempotent replays).
func (s *PaymentService) HandleNotify(ctx context.Context, provider string, rawBody []byte, headers map[string][]string, query map[string][]string) (statusCode int, body string, contentType string, err error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	prov, ok := s.providers[provider]
	if !ok || prov == nil {
		return 0, "", "", ErrPaymentProviderNotFound
	}

	ev, err := prov.VerifyAndParseNotify(ctx, rawBody, headers, query)
	if err != nil {
		return 0, "", "", err
	}

	// Best-effort: store callback payload for audit/idempotency.
	// Note: some failures below may still return non-2xx so provider retries; idempotency relies on DB uniqueness + order status checks.
	n := &PaymentNotification{
		Provider:   provider,
		EventID:    ev.EventID,
		RawBody:    string(rawBody),
		Verified:   true,
		Processed:  false,
		ReceivedAt: ev.Now,
	}
	if v := strings.TrimSpace(ev.OrderNo); v != "" {
		n.OrderNo = &v
	}
	if v := strings.TrimSpace(ev.ProviderTradeNo); v != "" {
		n.ProviderTradeNo = &v
	}
	if ev.AmountCents > 0 {
		v := ev.AmountCents
		n.AmountCents = &v
	}
	if v := strings.TrimSpace(ev.Currency); v != "" {
		n.Currency = &v
	}

	if _, err := s.notifyRepo.CreateIfNotExists(ctx, n); err != nil {
		return 0, "", "", fmt.Errorf("create payment notification: %w", err)
	}

	// Not paid yet: mark processed and return success (avoid retries).
	if !ev.Paid {
		_ = s.notifyRepo.MarkProcessed(ctx, provider, ev.EventID, true, true, nil)
		statusCode, body, contentType = prov.SuccessResponse()
		return statusCode, body, contentType, nil
	}

	// Process paid event in a DB transaction to ensure atomicity of state transitions and entitlement grants.
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return 0, "", "", fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	order, err := s.orderRepo.LockByOrderNo(txCtx, ev.OrderNo)
	if err != nil {
		if dbent.IsNotFound(err) {
			msg := "order not found"
			_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, &msg)
			if cerr := tx.Commit(); cerr != nil {
				return 0, "", "", fmt.Errorf("commit transaction: %w", cerr)
			}
			statusCode, body, contentType = prov.SuccessResponse()
			return statusCode, body, contentType, nil
		}
		return 0, "", "", fmt.Errorf("lock order: %w", err)
	}

	// Basic sanity checks.
	if strings.ToLower(strings.TrimSpace(order.Provider)) != provider {
		msg := "provider mismatch"
		_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, &msg)
		if cerr := tx.Commit(); cerr != nil {
			return 0, "", "", fmt.Errorf("commit transaction: %w", cerr)
		}
		statusCode, body, contentType = prov.SuccessResponse()
		return statusCode, body, contentType, nil
	}
	if strings.TrimSpace(order.Currency) != "" && strings.TrimSpace(ev.Currency) != "" &&
		!strings.EqualFold(strings.TrimSpace(order.Currency), strings.TrimSpace(ev.Currency)) {
		msg := "currency mismatch"
		_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, &msg)
		if cerr := tx.Commit(); cerr != nil {
			return 0, "", "", fmt.Errorf("commit transaction: %w", cerr)
		}
		statusCode, body, contentType = prov.SuccessResponse()
		return statusCode, body, contentType, nil
	}
	if ev.AmountCents > 0 && order.AmountCents != ev.AmountCents {
		msg := "amount mismatch"
		_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, &msg)
		if cerr := tx.Commit(); cerr != nil {
			return 0, "", "", fmt.Errorf("commit transaction: %w", cerr)
		}
		statusCode, body, contentType = prov.SuccessResponse()
		return statusCode, body, contentType, nil
	}

	// If order is cancelled/failed, do not fulfill; acknowledge to stop retries but keep audit.
	switch order.Status {
	case PaymentOrderStatusCancelled, PaymentOrderStatusFailed:
		msg := "order not fulfillable: " + order.Status
		_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, &msg)
		if cerr := tx.Commit(); cerr != nil {
			return 0, "", "", fmt.Errorf("commit transaction: %w", cerr)
		}
		statusCode, body, contentType = prov.SuccessResponse()
		return statusCode, body, contentType, nil
	}

	// Update paid fields (idempotent).
	now := ev.Now
	if order.ProviderTradeNo == nil && strings.TrimSpace(ev.ProviderTradeNo) != "" {
		v := strings.TrimSpace(ev.ProviderTradeNo)
		order.ProviderTradeNo = &v
	}
	if order.Status == PaymentOrderStatusCreated {
		order.Status = PaymentOrderStatusPaid
	}
	if order.PaidAt == nil {
		order.PaidAt = &now
	}
	if err := s.orderRepo.Update(txCtx, order); err != nil {
		return 0, "", "", fmt.Errorf("update order paid fields: %w", err)
	}

	// Fulfill once.
	if order.Status != PaymentOrderStatusFulfilled {
		if err := s.fulfillPaidOrder(txCtx, order, now); err != nil {
			msg := "fulfill error: " + err.Error()
			_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, false, true, &msg)
			return 0, "", "", ErrPaymentNotifyProcessError.WithCause(err)
		}
		order.Status = PaymentOrderStatusFulfilled
		order.FulfilledAt = &now
		if err := s.orderRepo.Update(txCtx, order); err != nil {
			return 0, "", "", fmt.Errorf("update order fulfilled: %w", err)
		}
	}

	// Mark notification processed in the same transaction.
	_ = s.notifyRepo.MarkProcessed(txCtx, provider, ev.EventID, true, true, nil)

	if err := tx.Commit(); err != nil {
		return 0, "", "", fmt.Errorf("commit transaction: %w", err)
	}

	// Cache invalidation after commit (best effort).
	s.invalidatePaymentCaches(ctx, order)

	statusCode, body, contentType = prov.SuccessResponse()
	return statusCode, body, contentType, nil
}

func (s *PaymentService) fulfillPaidOrder(ctx context.Context, order *PaymentOrder, now time.Time) error {
	if order == nil {
		return errors.New("nil order")
	}

	// Grant entitlements.
	switch order.Kind {
	case PaymentKindSubscription:
		if order.GrantGroupID == nil {
			return fmt.Errorf("missing grant_group_id")
		}
		validityDays := 30
		if order.GrantValidityDays != nil && *order.GrantValidityDays > 0 {
			validityDays = *order.GrantValidityDays
		}
		if s.subService == nil {
			return fmt.Errorf("nil subscription service")
		}
		if _, _, err := s.subService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
			UserID:       order.UserID,
			GroupID:      *order.GrantGroupID,
			ValidityDays: validityDays,
			AssignedBy:   0,
			Notes:        fmt.Sprintf("支付订单 %s", order.OrderNo),
		}); err != nil {
			return err
		}

	case PaymentKindBalance:
		if order.GrantCreditBalance == nil {
			return fmt.Errorf("missing grant_credit_balance")
		}
		if err := s.userRepo.UpdateBalance(ctx, order.UserID, *order.GrantCreditBalance); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported order kind: %s", order.Kind)
	}

	// Audit entitlement event (idempotency relies on order status; duplicates are prevented by order fulfillment gate).
	if s.entitlementRepo != nil {
		ev := &EntitlementEvent{
			UserID:    order.UserID,
			Kind:      order.Kind,
			Source:    "paid",
			OrderID:   func() *int64 { v := order.ID; return &v }(),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if order.GrantGroupID != nil {
			v := *order.GrantGroupID
			ev.GroupID = &v
		}
		if order.GrantValidityDays != nil {
			v := *order.GrantValidityDays
			ev.ValidityDays = &v
		}
		if order.GrantCreditBalance != nil {
			v := *order.GrantCreditBalance
			ev.BalanceDelta = &v
		}
		note := fmt.Sprintf("支付回调发放（%s）", order.Provider)
		ev.Note = &note
		if err := s.entitlementRepo.Create(ctx, ev); err != nil {
			return fmt.Errorf("create entitlement event: %w", err)
		}
	}

	return nil
}

func (s *PaymentService) invalidatePaymentCaches(ctx context.Context, order *PaymentOrder) {
	if order == nil {
		return
	}
	if s.authInvalidator != nil {
		s.authInvalidator.InvalidateAuthCacheByUserID(ctx, order.UserID)
	}
	if s.billingCache == nil {
		return
	}

	switch order.Kind {
	case PaymentKindBalance:
		go func(userID int64) {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCache.InvalidateUserBalance(cacheCtx, userID)
		}(order.UserID)
	case PaymentKindSubscription:
		if order.GrantGroupID == nil {
			return
		}
		userID, groupID := order.UserID, *order.GrantGroupID
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCache.InvalidateSubscription(cacheCtx, userID, groupID)
		}()
	}
}
