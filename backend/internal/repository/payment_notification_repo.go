package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentnotification"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentNotificationRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewPaymentNotificationRepository(client *dbent.Client, db *sql.DB) service.PaymentNotificationRepository {
	return &paymentNotificationRepository{client: client, db: db}
}

func (r *paymentNotificationRepository) CreateIfNotExists(ctx context.Context, n *service.PaymentNotification) (bool, error) {
	// 优先使用原生 SQL：INSERT ... ON CONFLICT DO NOTHING RETURNING id
	// 这样可以准确区分 inserted=true/false，用于回调幂等处理。
	if r.db != nil {
		query := `
INSERT INTO payment_notifications
  (provider, event_id, order_no, provider_trade_no, amount_cents, currency,
   verified, processed, process_error, raw_body, received_at)
VALUES
  ($1, $2, $3, $4, $5, $6,
   $7, $8, $9, $10, COALESCE($11, NOW()))
ON CONFLICT (provider, event_id) DO NOTHING
RETURNING id;
`

		var (
			orderNo         any
			providerTradeNo any
			amountCents     any
			currency        any
			processError    any
			receivedAt      any
		)
		if n.OrderNo != nil {
			orderNo = *n.OrderNo
		}
		if n.ProviderTradeNo != nil {
			providerTradeNo = *n.ProviderTradeNo
		}
		if n.AmountCents != nil {
			amountCents = *n.AmountCents
		}
		if n.Currency != nil {
			currency = *n.Currency
		}
		if n.ProcessError != nil {
			pe := strings.TrimSpace(*n.ProcessError)
			if pe != "" {
				processError = pe
			}
		}
		if !n.ReceivedAt.IsZero() {
			receivedAt = n.ReceivedAt
		}

		var id int64
		err := r.db.QueryRowContext(
			ctx,
			query,
			n.Provider,
			n.EventID,
			orderNo,
			providerTradeNo,
			amountCents,
			currency,
			n.Verified,
			n.Processed,
			processError,
			n.RawBody,
			receivedAt,
		).Scan(&id)
		if err == nil {
			n.ID = id
			return true, nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	// fallback：使用 Ent（无法准确区分 inserted=false，只能 best-effort）
	client := clientFromContext(ctx, r.client)
	err := client.PaymentNotification.Create().
		SetProvider(n.Provider).
		SetEventID(n.EventID).
		SetRawBody(n.RawBody).
		OnConflictColumns(paymentnotification.FieldProvider, paymentnotification.FieldEventID).
		DoNothing().
		Exec(ctx)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (r *paymentNotificationRepository) MarkProcessed(ctx context.Context, provider, eventID string, processed bool, verified bool, processError *string) error {
	client := clientFromContext(ctx, r.client)
	b := client.PaymentNotification.Update().
		Where(
			paymentnotification.ProviderEQ(provider),
			paymentnotification.EventIDEQ(eventID),
		).
		SetProcessed(processed).
		SetVerified(verified)

	if processError != nil {
		errText := strings.TrimSpace(*processError)
		if errText != "" {
			b.SetProcessError(errText)
		} else {
			b.ClearProcessError()
		}
	}

	_, err := b.Save(ctx)
	return err
}

func (r *paymentNotificationRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, provider, orderNo, search string) ([]service.PaymentNotification, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.PaymentNotification.Query()

	if s := strings.TrimSpace(provider); s != "" {
		q = q.Where(paymentnotification.ProviderEQ(s))
	}
	if s := strings.TrimSpace(orderNo); s != "" {
		q = q.Where(paymentnotification.OrderNoEQ(s))
	}
	if s := strings.TrimSpace(search); s != "" {
		q = q.Where(
			paymentnotification.Or(
				paymentnotification.EventIDContainsFold(s),
				paymentnotification.ProviderTradeNoContainsFold(s),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	q = q.Order(dbent.Desc(paymentnotification.FieldReceivedAt)).
		Offset(params.Offset()).
		Limit(params.Limit())
	rows, err := q.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	items := make([]service.PaymentNotification, 0, len(rows))
	for i := range rows {
		items = append(items, *paymentNotificationEntityToService(rows[i]))
	}

	pages := int((int64(total) + int64(params.Limit()) - 1) / int64(params.Limit()))
	if pages < 1 {
		pages = 1
	}
	return items, &pagination.PaginationResult{Total: int64(total), Page: params.Page, PageSize: params.PageSize, Pages: pages}, nil
}

func paymentNotificationEntityToService(m *dbent.PaymentNotification) *service.PaymentNotification {
	if m == nil {
		return nil
	}
	out := &service.PaymentNotification{
		ID:        m.ID,
		Provider:  m.Provider,
		EventID:   m.EventID,
		Verified:  m.Verified,
		Processed: m.Processed,
		RawBody:   m.RawBody,
		ReceivedAt: m.ReceivedAt,
	}

	if m.OrderNo != nil {
		v := *m.OrderNo
		out.OrderNo = &v
	}
	if m.ProviderTradeNo != nil {
		v := *m.ProviderTradeNo
		out.ProviderTradeNo = &v
	}
	if m.AmountCents != nil {
		v := *m.AmountCents
		out.AmountCents = &v
	}
	if m.Currency != nil {
		v := *m.Currency
		out.Currency = &v
	}
	if m.ProcessError != nil {
		v := *m.ProcessError
		out.ProcessError = &v
	}

	return out
}
