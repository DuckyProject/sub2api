package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type epayProvider struct {
	settingService *SettingService
}

func (p *epayProvider) Name() string { return PaymentProviderEpay }

func (p *epayProvider) CreatePayment(ctx context.Context, order *PaymentOrder) (string, string, error) {
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
	if !settings.PaymentEnabled || !settings.PaymentEpayEnabled ||
		strings.TrimSpace(settings.PaymentEpayGatewayURL) == "" ||
		strings.TrimSpace(settings.PaymentEpayPID) == "" ||
		strings.TrimSpace(settings.PaymentEpayKey) == "" {
		return "", "", ErrPaymentProviderMisconfigured
	}

	notifyURL, err := buildNotifyURL(settings.PublicBaseURL, PaymentProviderEpay)
	if err != nil {
		return "", "", err
	}
	returnURL := buildReturnURL(settings.PublicBaseURL)

	gatewayURL, err := normalizeEpayGatewayURL(settings.PaymentEpayGatewayURL)
	if err != nil {
		return "", "", err
	}

	params := map[string]string{
		"pid":          strings.TrimSpace(settings.PaymentEpayPID),
		"out_trade_no": strings.TrimSpace(order.OrderNo),
		"notify_url":   notifyURL,
		"name":         buildOrderSubject(order),
		"money":        formatCents(order.AmountCents),
	}
	if returnURL != "" {
		params["return_url"] = returnURL
	}

	signBase := buildSortedQueryString(params, nil)
	sign := md5Hex(signBase + strings.TrimSpace(settings.PaymentEpayKey))

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

	// Best-effort expiry hint for internal bookkeeping.
	if order.ExpiresAt == nil {
		exp := time.Now().Add(30 * time.Minute)
		order.ExpiresAt = &exp
	}
	return payURL, "", nil
}

func (p *epayProvider) VerifyAndParseNotify(ctx context.Context, rawBody []byte, headers map[string][]string, query map[string][]string) (*ProviderNotifyEvent, error) {
	_ = headers
	if p.settingService == nil {
		return nil, ErrPaymentProviderMisconfigured
	}

	settings, err := p.settingService.GetAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if strings.TrimSpace(settings.PaymentEpayKey) == "" || strings.TrimSpace(settings.PaymentEpayPID) == "" {
		return nil, ErrPaymentProviderMisconfigured
	}

	params := parseNotifyParams(rawBody, query)

	// Optional: verify pid matches.
	if pid := strings.TrimSpace(params["pid"]); pid != "" && pid != strings.TrimSpace(settings.PaymentEpayPID) {
		return nil, ErrPaymentSignatureInvalid
	}

	ok, _ := verifyMD5Signature(params, strings.TrimSpace(settings.PaymentEpayKey))
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

	money := strings.TrimSpace(params["money"])
	if money == "" {
		money = strings.TrimSpace(params["amount"])
	}
	amountCents, err := parseMoneyToCents(money)
	if err != nil {
		return nil, ErrPaymentNotifyInvalid.WithCause(err)
	}

	paid, okPaid := parsePaidFlag(params)
	if !okPaid {
		return nil, ErrPaymentNotifyInvalid
	}

	eventID := tradeNo
	if eventID == "" {
		// Fallback: stable per orderNo (some gateways omit trade_no)
		eventID = orderNo
	}

	return &ProviderNotifyEvent{
		Provider:        PaymentProviderEpay,
		EventID:         eventID,
		OrderNo:         orderNo,
		ProviderTradeNo: tradeNo,
		AmountCents:     amountCents,
		Currency:        "CNY",
		Paid:            paid,
		RawBody:         string(rawBody),
		Now:             nowForNotify(ctx),
	}, nil
}

func (p *epayProvider) SuccessResponse() (int, string, string) {
	return 200, "success", "text/plain; charset=utf-8"
}

func normalizeEpayGatewayURL(raw string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return "", ErrPaymentProviderMisconfigured
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", ErrPaymentProviderMisconfigured.WithCause(fmt.Errorf("invalid gateway url: %q", raw))
	}

	// Many Epay deployments expose submit.php at the gateway root.
	// If admin configured a full script URL already, keep it as-is.
	path := strings.TrimSpace(u.Path)
	if path == "" || strings.HasSuffix(path, "/") {
		u.Path = strings.TrimRight(path, "/") + "/submit.php"
	}
	u.RawQuery = strings.TrimLeft(u.RawQuery, "?")
	u.Fragment = ""
	return strings.TrimRight(u.String(), "?"), nil
}

func buildOrderSubject(order *PaymentOrder) string {
	if order == nil {
		return "Sub2API"
	}
	switch order.Kind {
	case PaymentKindSubscription:
		return "Subscription"
	case PaymentKindBalance:
		return "Balance"
	default:
		return "Sub2API"
	}
}

// compile-time check
var _ PaymentProvider = (*epayProvider)(nil)
