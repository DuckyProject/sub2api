package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_BillingRateMultiplier_DefaultsToOneWhenNil(t *testing.T) {
	var a Account
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"name":"acc","status":"active"}`), &a))
	require.Nil(t, a.RateMultiplier)
	require.Equal(t, 1.0, a.BillingRateMultiplier())
}

func TestAccount_BillingRateMultiplier_AllowsZero(t *testing.T) {
	a := Account{RateMultiplier: new(0.0)}
	require.Equal(t, 0.0, a.BillingRateMultiplier())
}

func TestAccount_BillingRateMultiplier_NegativeFallsBackToOne(t *testing.T) {
	a := Account{RateMultiplier: new(-1.0)}
	require.Equal(t, 1.0, a.BillingRateMultiplier())
}
