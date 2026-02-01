package repository

import (
	"context"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentOrderRepository struct {
	client *dbent.Client
}

func NewPaymentOrderRepository(client *dbent.Client) service.PaymentOrderRepository {
	return &paymentOrderRepository{client: client}
}

func (r *paymentOrderRepository) Create(ctx context.Context, o *service.PaymentOrder) error {
	client := clientFromContext(ctx, r.client)
	b := client.PaymentOrder.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetKind(o.Kind).
		SetStatus(o.Status).
		SetProvider(o.Provider).
		SetCurrency(o.Currency).
		SetAmountCents(o.AmountCents)

	if o.ProductID != nil {
		b.SetProductID(*o.ProductID)
	}
	if o.ClientRequestID != nil {
		b.SetClientRequestID(*o.ClientRequestID)
	}
	if o.ProviderTradeNo != nil {
		b.SetProviderTradeNo(*o.ProviderTradeNo)
	}
	if o.PayURL != nil {
		b.SetPayURL(*o.PayURL)
	}
	if o.ExpiresAt != nil {
		b.SetExpiresAt(*o.ExpiresAt)
	}
	if o.PaidAt != nil {
		b.SetPaidAt(*o.PaidAt)
	}
	if o.FulfilledAt != nil {
		b.SetFulfilledAt(*o.FulfilledAt)
	}
	if o.GrantGroupID != nil {
		b.SetGrantGroupID(*o.GrantGroupID)
	}
	if o.GrantValidityDays != nil {
		b.SetGrantValidityDays(*o.GrantValidityDays)
	}
	if o.GrantCreditBalance != nil {
		b.SetGrantCreditBalance(*o.GrantCreditBalance)
	}
	if o.Notes != nil {
		b.SetNotes(*o.Notes)
	}

	m, err := b.Save(ctx)
	if err != nil {
		return err
	}
	applyPaymentOrderEntityToService(o, m)
	return nil
}

func (r *paymentOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Query().Where(paymentorder.OrderNoEQ(orderNo)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) GetByProviderTradeNo(ctx context.Context, provider, tradeNo string) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Query().
		Where(
			paymentorder.ProviderEQ(provider),
			paymentorder.ProviderTradeNoEQ(tradeNo),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) GetByUserAndClientRequestID(ctx context.Context, userID int64, clientRequestID string) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(userID),
			paymentorder.ClientRequestIDEQ(clientRequestID),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) LockByOrderNo(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Query().
		Where(paymentorder.OrderNoEQ(orderNo)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) Update(ctx context.Context, o *service.PaymentOrder) error {
	client := clientFromContext(ctx, r.client)
	b := client.PaymentOrder.UpdateOneID(o.ID).
		SetStatus(o.Status).
		SetProvider(o.Provider).
		SetCurrency(o.Currency).
		SetAmountCents(o.AmountCents)

	if o.ProductID != nil {
		b.SetProductID(*o.ProductID)
	} else {
		b.ClearProductID()
	}
	if o.ClientRequestID != nil {
		b.SetClientRequestID(*o.ClientRequestID)
	} else {
		b.ClearClientRequestID()
	}
	if o.ProviderTradeNo != nil {
		b.SetProviderTradeNo(*o.ProviderTradeNo)
	} else {
		b.ClearProviderTradeNo()
	}
	if o.PayURL != nil {
		b.SetPayURL(*o.PayURL)
	} else {
		b.ClearPayURL()
	}
	if o.ExpiresAt != nil {
		b.SetExpiresAt(*o.ExpiresAt)
	} else {
		b.ClearExpiresAt()
	}
	if o.PaidAt != nil {
		b.SetPaidAt(*o.PaidAt)
	} else {
		b.ClearPaidAt()
	}
	if o.FulfilledAt != nil {
		b.SetFulfilledAt(*o.FulfilledAt)
	} else {
		b.ClearFulfilledAt()
	}
	if o.GrantGroupID != nil {
		b.SetGrantGroupID(*o.GrantGroupID)
	} else {
		b.ClearGrantGroupID()
	}
	if o.GrantValidityDays != nil {
		b.SetGrantValidityDays(*o.GrantValidityDays)
	} else {
		b.ClearGrantValidityDays()
	}
	if o.GrantCreditBalance != nil {
		b.SetGrantCreditBalance(*o.GrantCreditBalance)
	} else {
		b.ClearGrantCreditBalance()
	}
	if o.Notes != nil {
		b.SetNotes(*o.Notes)
	} else {
		b.ClearNotes()
	}

	m, err := b.Save(ctx)
	if err != nil {
		return err
	}
	o.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *paymentOrderRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	uid := userID
	return r.ListWithFilters(ctx, params, &uid, "", status, "", "")
}

func (r *paymentOrderRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, userID *int64, kind, status, provider, search string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.PaymentOrder.Query()

	if userID != nil && *userID > 0 {
		q = q.Where(paymentorder.UserIDEQ(*userID))
	}
	if s := strings.TrimSpace(kind); s != "" {
		q = q.Where(paymentorder.KindEQ(s))
	}
	if s := strings.TrimSpace(status); s != "" {
		q = q.Where(paymentorder.StatusEQ(s))
	}
	if s := strings.TrimSpace(provider); s != "" {
		q = q.Where(paymentorder.ProviderEQ(s))
	}
	if s := strings.TrimSpace(search); s != "" {
		q = q.Where(
			paymentorder.Or(
				paymentorder.OrderNoContainsFold(s),
				paymentorder.ProviderTradeNoContainsFold(s),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	q = q.Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit())
	rows, err := q.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.PaymentOrder, 0, len(rows))
	for i := range rows {
		items = append(items, *paymentOrderEntityToService(rows[i]))
	}

	pages := int((int64(total) + int64(params.Limit()) - 1) / int64(params.Limit()))
	if pages < 1 {
		pages = 1
	}
	return items, &pagination.PaginationResult{Total: int64(total), Page: params.Page, PageSize: params.PageSize, Pages: pages}, nil
}

func paymentOrderEntityToService(m *dbent.PaymentOrder) *service.PaymentOrder {
	if m == nil {
		return nil
	}
	out := &service.PaymentOrder{}
	applyPaymentOrderEntityToService(out, m)
	return out
}

func applyPaymentOrderEntityToService(out *service.PaymentOrder, m *dbent.PaymentOrder) {
	out.ID = m.ID
	out.OrderNo = m.OrderNo
	out.UserID = m.UserID
	out.Kind = m.Kind
	out.Status = m.Status
	out.Provider = m.Provider
	out.Currency = m.Currency
	out.AmountCents = m.AmountCents
	out.CreatedAt = m.CreatedAt
	out.UpdatedAt = m.UpdatedAt

	out.ProductID = m.ProductID
	out.ClientRequestID = m.ClientRequestID
	out.ProviderTradeNo = m.ProviderTradeNo
	out.PayURL = m.PayURL
	out.ExpiresAt = m.ExpiresAt
	out.PaidAt = m.PaidAt
	out.FulfilledAt = m.FulfilledAt
	out.GrantGroupID = m.GrantGroupID
	out.GrantValidityDays = m.GrantValidityDays
	out.GrantCreditBalance = m.GrantCreditBalance
	out.Notes = m.Notes

	// Ensure non-nil timestamps in service layer when needed.
	if out.CreatedAt.IsZero() {
		out.CreatedAt = time.Now()
	}
}
