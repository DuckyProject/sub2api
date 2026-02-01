package repository

import (
	"context"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentproduct"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentProductRepository struct {
	client *dbent.Client
}

func NewPaymentProductRepository(client *dbent.Client) service.PaymentProductRepository {
	return &paymentProductRepository{client: client}
}

func (r *paymentProductRepository) Create(ctx context.Context, p *service.PaymentProduct) error {
	client := clientFromContext(ctx, r.client)
	b := client.PaymentProduct.Create().
		SetKind(p.Kind).
		SetName(p.Name).
		SetDescriptionMd(p.DescriptionMD).
		SetStatus(p.Status).
		SetSortOrder(p.SortOrder).
		SetCurrency(p.Currency).
		SetPriceCents(p.PriceCents).
		SetAllowCustomAmount(p.AllowCustomAmount)

	if p.GroupID != nil {
		b.SetGroupID(*p.GroupID)
	}
	if p.ValidityDays != nil {
		b.SetValidityDays(*p.ValidityDays)
	}
	if p.CreditBalance != nil {
		b.SetCreditBalance(*p.CreditBalance)
	}
	if p.MinAmountCents != nil {
		b.SetMinAmountCents(*p.MinAmountCents)
	}
	if p.MaxAmountCents != nil {
		b.SetMaxAmountCents(*p.MaxAmountCents)
	}
	if p.SuggestedAmountsCents == nil {
		b.SetSuggestedAmountsCents([]int64{})
	} else {
		b.SetSuggestedAmountsCents(p.SuggestedAmountsCents)
	}
	if p.ExchangeRate != nil {
		b.SetExchangeRate(*p.ExchangeRate)
	}

	m, err := b.Save(ctx)
	if err != nil {
		return err
	}
	p.ID = m.ID
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *paymentProductRepository) Update(ctx context.Context, p *service.PaymentProduct) error {
	client := clientFromContext(ctx, r.client)
	b := client.PaymentProduct.UpdateOneID(p.ID).
		SetKind(p.Kind).
		SetName(p.Name).
		SetDescriptionMd(p.DescriptionMD).
		SetStatus(p.Status).
		SetSortOrder(p.SortOrder).
		SetCurrency(p.Currency).
		SetPriceCents(p.PriceCents).
		SetAllowCustomAmount(p.AllowCustomAmount)

	if p.GroupID != nil {
		b.SetGroupID(*p.GroupID)
	} else {
		b.ClearGroupID()
	}
	if p.ValidityDays != nil {
		b.SetValidityDays(*p.ValidityDays)
	} else {
		b.ClearValidityDays()
	}
	if p.CreditBalance != nil {
		b.SetCreditBalance(*p.CreditBalance)
	} else {
		b.ClearCreditBalance()
	}
	if p.MinAmountCents != nil {
		b.SetMinAmountCents(*p.MinAmountCents)
	} else {
		b.ClearMinAmountCents()
	}
	if p.MaxAmountCents != nil {
		b.SetMaxAmountCents(*p.MaxAmountCents)
	} else {
		b.ClearMaxAmountCents()
	}
	b.SetSuggestedAmountsCents(p.SuggestedAmountsCents)
	if p.ExchangeRate != nil {
		b.SetExchangeRate(*p.ExchangeRate)
	} else {
		b.ClearExchangeRate()
	}

	m, err := b.Save(ctx)
	if err != nil {
		return err
	}
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *paymentProductRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.PaymentProduct.DeleteOneID(id).Exec(ctx)
}

func (r *paymentProductRepository) GetByID(ctx context.Context, id int64) (*service.PaymentProduct, error) {
	m, err := r.client.PaymentProduct.Query().Where(paymentproduct.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return paymentProductEntityToService(m), nil
}

func (r *paymentProductRepository) ListActiveByKind(ctx context.Context, kind string) ([]service.PaymentProduct, error) {
	rows, err := r.client.PaymentProduct.Query().
		Where(
			paymentproduct.StatusEQ(service.PaymentProductStatusActive),
			paymentproduct.KindEQ(kind),
		).
		Order(dbent.Asc(paymentproduct.FieldSortOrder), dbent.Asc(paymentproduct.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]service.PaymentProduct, 0, len(rows))
	for i := range rows {
		out = append(out, *paymentProductEntityToService(rows[i]))
	}
	return out, nil
}

func (r *paymentProductRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, kind, status, search string) ([]service.PaymentProduct, *pagination.PaginationResult, error) {
	q := r.client.PaymentProduct.Query()
	if strings.TrimSpace(kind) != "" {
		q = q.Where(paymentproduct.KindEQ(strings.TrimSpace(kind)))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where(paymentproduct.StatusEQ(strings.TrimSpace(status)))
	}
	if s := strings.TrimSpace(search); s != "" {
		q = q.Where(paymentproduct.NameContainsFold(s))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	q = q.Order(dbent.Desc(paymentproduct.FieldCreatedAt))
	q = q.Offset(params.Offset()).Limit(params.Limit())

	rows, err := q.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.PaymentProduct, 0, len(rows))
	for i := range rows {
		items = append(items, *paymentProductEntityToService(rows[i]))
	}
	pages := int((int64(total) + int64(params.Limit()) - 1) / int64(params.Limit()))
	if pages < 1 {
		pages = 1
	}
	return items, &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func paymentProductEntityToService(m *dbent.PaymentProduct) *service.PaymentProduct {
	if m == nil {
		return nil
	}
	out := &service.PaymentProduct{
		ID:                    m.ID,
		Kind:                  m.Kind,
		Name:                  m.Name,
		DescriptionMD:         m.DescriptionMd,
		Status:                m.Status,
		SortOrder:             m.SortOrder,
		Currency:              m.Currency,
		PriceCents:            m.PriceCents,
		AllowCustomAmount:     m.AllowCustomAmount,
		SuggestedAmountsCents: m.SuggestedAmountsCents,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
	if m.GroupID != nil {
		v := *m.GroupID
		out.GroupID = &v
	}
	if m.ValidityDays != nil {
		v := *m.ValidityDays
		out.ValidityDays = &v
	}
	if m.CreditBalance != nil {
		v := *m.CreditBalance
		out.CreditBalance = &v
	}
	if m.MinAmountCents != nil {
		v := *m.MinAmountCents
		out.MinAmountCents = &v
	}
	if m.MaxAmountCents != nil {
		v := *m.MaxAmountCents
		out.MaxAmountCents = &v
	}
	if m.ExchangeRate != nil {
		v := *m.ExchangeRate
		out.ExchangeRate = &v
	}
	return out
}
