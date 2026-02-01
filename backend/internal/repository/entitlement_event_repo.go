package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type entitlementEventRepository struct {
	client *dbent.Client
}

func NewEntitlementEventRepository(client *dbent.Client) service.EntitlementEventRepository {
	return &entitlementEventRepository{client: client}
}

func (r *entitlementEventRepository) Create(ctx context.Context, e *service.EntitlementEvent) error {
	client := clientFromContext(ctx, r.client)
	b := client.EntitlementEvent.Create().
		SetUserID(e.UserID).
		SetKind(e.Kind).
		SetSource(e.Source)

	if e.GroupID != nil {
		b.SetGroupID(*e.GroupID)
	}
	if e.ValidityDays != nil {
		b.SetValidityDays(*e.ValidityDays)
	}
	if e.BalanceDelta != nil {
		b.SetBalanceDelta(*e.BalanceDelta)
	}
	if e.ConcurrencyDelta != nil {
		b.SetConcurrencyDelta(*e.ConcurrencyDelta)
	}
	if e.OrderID != nil {
		b.SetOrderID(*e.OrderID)
	}
	if e.RedeemCodeID != nil {
		b.SetRedeemCodeID(*e.RedeemCodeID)
	}
	if e.ActorUserID != nil {
		b.SetActorUserID(*e.ActorUserID)
	}
	if e.Note != nil {
		b.SetNote(*e.Note)
	}

	m, err := b.Save(ctx)
	if err != nil {
		return err
	}
	e.ID = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return nil
}
