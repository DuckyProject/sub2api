package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type tokenPayProvider struct {
	settingService *SettingService
}

func (p *tokenPayProvider) Name() string { return PaymentProviderTokenPay }

func (p *tokenPayProvider) CreatePayment(ctx context.Context, order *PaymentOrder) (string, string, error) {
	if order == nil {
		return "", "", ErrPaymentNotifyInvalid
	}
	if p.settingService == nil {
		return "", "", ErrPaymentProviderMisconfigured
	}

	settings, err := p.settingService.GetAllSettings(ctx)
	if err != nil {
		return "", "", fmt.Errorf("get settings: %w", err)
	}
	if !settings.PaymentEnabled || !settings.PaymentTokenPayEnabled ||
		strings.TrimSpace(settings.PaymentTokenPayGatewayURL) == "" ||
		strings.TrimSpace(settings.PaymentTokenPayMerchantID) == "" ||
		strings.TrimSpace(settings.PaymentTokenPayKey) == "" {
		return "", "", ErrPaymentProviderMisconfigured
	}

	notifyURL, err := buildNotifyURL(settings.PublicBaseURL, PaymentProviderTokenPay)
	if err != nil {
		return "", "", err
	}
	returnURL := buildReturnURL(settings.PublicBaseURL)

	gatewayURL, err := normalizeGenericGatewayURL(settings.PaymentTokenPayGatewayURL)
	if err != nil {
		return "", "", err
	}

	params := map[string]string{
		"merchant_id":  strings.TrimSpace(settings.PaymentTokenPayMerchantID),
		"out_trade_no": strings.TrimSpace(order.OrderNo),
		"notify_url":   notifyURL,
		"name":         buildOrderSubject(order),
		"money":        formatCents(order.AmountCents),
	}
	if returnURL != "" {
		params["return_url"] = returnURL
	}
	// Optional: propagate currency if product is non-CNY.
	if c := strings.TrimSpace(order.Currency); c != "" && strings.ToUpper(c) != "CNY" {
		params["currency"] = c
	}

	signBase := buildSortedQueryString(params, nil)
	sign := md5Hex(signBase + strings.TrimSpace(settings.PaymentTokenPayKey))

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("sign", sign)
	q.Set("sign_type", "MD5")

	payURL := gatewayURL
	if strings.Contains(payURL, "?") {
		payURL += "&" + q.Encode()
	} else {
		payURL += "?" + q.Encode()
	}

	if order.ExpiresAt == nil {
		exp := time.Now().Add(30 * time.Minute)
		order.ExpiresAt = &exp
	}
	return payURL, "", nil
}

func (p *tokenPayProvider) VerifyAndParseNotify(ctx context.Context, rawBody []byte, headers map[string][]string, query map[string][]string) (*ProviderNotifyEvent, error) {
	_ = headers
	if p.settingService == nil {
		return nil, ErrPaymentProviderMisconfigured
	}

	settings, err := p.settingService.GetAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if strings.TrimSpace(settings.PaymentTokenPayKey) == "" || strings.TrimSpace(settings.PaymentTokenPayMerchantID) == "" {
		return nil, ErrPaymentProviderMisconfigured
	}

	params := parseNotifyParams(rawBody, query)

	// Optional: verify merchant id matches.
	if mid := strings.TrimSpace(params["merchant_id"]); mid != "" && mid != strings.TrimSpace(settings.PaymentTokenPayMerchantID) {
		return nil, ErrPaymentSignatureInvalid
	}

	ok, _ := verifyMD5Signature(params, strings.TrimSpace(settings.PaymentTokenPayKey))
	if !ok {
		return nil, ErrPaymentSignatureInvalid
	}

	orderNo := strings.TrimSpace(params["out_trade_no"])
	if orderNo == "" {
		orderNo = strings.TrimSpace(params["order_no"])
	}
	if orderNo == "" {
		return nil, ErrPaymentNotifyInvalid
	}

	tradeNo := strings.TrimSpace(params["trade_no"])
	if tradeNo == "" {
		tradeNo = strings.TrimSpace(params["transaction_id"])
	}

	amountRaw := strings.TrimSpace(params["money"])
	if amountRaw == "" {
		amountRaw = strings.TrimSpace(params["amount"])
	}
	amountCents, err := parseMoneyToCents(amountRaw)
	if err != nil {
		return nil, ErrPaymentNotifyInvalid.WithCause(err)
	}

	paid, okPaid := parsePaidFlag(params)
	if !okPaid {
		return nil, ErrPaymentNotifyInvalid
	}

	currency := strings.TrimSpace(params["currency"])
	if currency == "" {
		currency = "CNY"
	}

	eventID := tradeNo
	if eventID == "" {
		eventID = orderNo
	}

	return &ProviderNotifyEvent{
		Provider:        PaymentProviderTokenPay,
		EventID:         eventID,
		OrderNo:         orderNo,
		ProviderTradeNo: tradeNo,
		AmountCents:     amountCents,
		Currency:        currency,
		Paid:            paid,
		RawBody:         string(rawBody),
		Now:             nowForNotify(ctx),
	}, nil
}

func (p *tokenPayProvider) SuccessResponse() (int, string, string) {
	return 200, "success", "text/plain; charset=utf-8"
}

func normalizeGenericGatewayURL(raw string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return "", ErrPaymentProviderMisconfigured
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", ErrPaymentProviderMisconfigured.WithCause(fmt.Errorf("invalid gateway url: %q", raw))
	}
	u.RawQuery = strings.TrimLeft(u.RawQuery, "?")
	u.Fragment = ""
	return strings.TrimRight(u.String(), "?"), nil
}

// compile-time check
var _ PaymentProvider = (*tokenPayProvider)(nil)
