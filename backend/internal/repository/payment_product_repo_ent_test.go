package repository

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newPaymentProductEntRepo(t *testing.T) (service.PaymentProductRepository, *dbent.Client) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_product_repo?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	return NewPaymentProductRepository(client), client
}

func TestPaymentProductRepositoryEntCreateSubscriptionProductPersistsGrantFields(t *testing.T) {
	repo, client := newPaymentProductEntRepo(t)

	ctx := context.Background()
	g, err := client.Group.Create().SetName("Test Subscription Group").Save(ctx)
	require.NoError(t, err)

	gid := g.ID
	validity := 30
	p := &service.PaymentProduct{
		Kind:          service.PaymentKindSubscription,
		Name:          "Monthly Plan",
		DescriptionMD: "test",
		Status:        service.PaymentProductStatusActive,
		SortOrder:     10,
		Currency:      "CNY",
		PriceCents:    9900,
		GroupID:       &gid,
		ValidityDays:  &validity,
		// Intentionally nil: repository should persist [] not NULL.
		SuggestedAmountsCents: nil,
	}

	require.NoError(t, repo.Create(ctx, p))
	require.NotZero(t, p.ID)

	loaded, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.GroupID)
	require.Equal(t, gid, *loaded.GroupID)
	require.NotNil(t, loaded.ValidityDays)
	require.Equal(t, validity, *loaded.ValidityDays)
	require.NotNil(t, loaded.SuggestedAmountsCents)
	require.Empty(t, loaded.SuggestedAmountsCents)
}

func TestPaymentProductRepositoryEntCreateBalanceProductPersistsConfigFields(t *testing.T) {
	repo, _ := newPaymentProductEntRepo(t)

	min := int64(100)   // 1.00
	max := int64(10000) // 100.00
	exchangeRate := 2.5
	creditBalance := 123.456

	p := &service.PaymentProduct{
		Kind:          service.PaymentKindBalance,
		Name:          "Balance Topup",
		DescriptionMD: "test",
		Status:        service.PaymentProductStatusActive,
		SortOrder:     1,
		Currency:      "CNY",
		PriceCents:    5000,

		AllowCustomAmount:     true,
		MinAmountCents:        &min,
		MaxAmountCents:        &max,
		SuggestedAmountsCents: []int64{100, 500, 1000},
		ExchangeRate:          &exchangeRate,
		CreditBalance:         &creditBalance,
	}

	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, p))
	require.NotZero(t, p.ID)

	loaded, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, service.PaymentKindBalance, loaded.Kind)
	require.Equal(t, p.Name, loaded.Name)
	require.Equal(t, p.Currency, loaded.Currency)
	require.Equal(t, p.PriceCents, loaded.PriceCents)
	require.Equal(t, p.AllowCustomAmount, loaded.AllowCustomAmount)

	require.NotNil(t, loaded.MinAmountCents)
	require.Equal(t, min, *loaded.MinAmountCents)
	require.NotNil(t, loaded.MaxAmountCents)
	require.Equal(t, max, *loaded.MaxAmountCents)
	require.Equal(t, []int64{100, 500, 1000}, loaded.SuggestedAmountsCents)

	require.NotNil(t, loaded.ExchangeRate)
	require.InDelta(t, exchangeRate, *loaded.ExchangeRate, 1e-9)
	require.NotNil(t, loaded.CreditBalance)
	require.InDelta(t, creditBalance, *loaded.CreditBalance, 1e-9)
}
