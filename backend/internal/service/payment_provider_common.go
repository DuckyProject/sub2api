package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrPaymentProviderMisconfigured = infraerrors.ServiceUnavailable("PAYMENT_PROVIDER_MISCONFIGURED", "payment provider is misconfigured")
	ErrPaymentPublicBaseURLRequired = infraerrors.BadRequest("PAYMENT_PUBLIC_BASE_URL_REQUIRED", "public_base_url is required for payment callbacks")
	ErrPaymentNotifyInvalid         = infraerrors.BadRequest("PAYMENT_NOTIFY_INVALID", "invalid payment notification")
	ErrPaymentSignatureInvalid      = infraerrors.BadRequest("PAYMENT_SIGNATURE_INVALID", "invalid payment signature")
)

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func buildSortedQueryString(params map[string]string, excludeKeys map[string]struct{}) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		if excludeKeys != nil {
			if _, ok := excludeKeys[k]; ok {
				continue
			}
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := strings.TrimSpace(params[k])
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}

func verifyMD5Signature(params map[string]string, key string) (bool, string) {
	sign := strings.TrimSpace(params["sign"])
	if sign == "" {
		return false, ""
	}

	exclude := map[string]struct{}{
		"sign":      {},
		"sign_type": {},
	}
	base := buildSortedQueryString(params, exclude)
	if base == "" {
		return false, ""
	}

	expect1 := md5Hex(base + key)
	if strings.EqualFold(sign, expect1) {
		return true, expect1
	}
	expect2 := md5Hex(base + "&key=" + key)
	if strings.EqualFold(sign, expect2) {
		return true, expect2
	}
	return false, expect1
}

func parseMoneyToCents(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty money")
	}
	// Normalize leading '+'.
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("negative money")
	}

	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	if intPart == "" {
		intPart = "0"
	}
	i, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money int part: %w", err)
	}

	var frac string
	if len(parts) == 2 {
		frac = parts[1]
	}
	// Allow up to 2 decimals (or more if all extra digits are zeros).
	if len(frac) > 2 {
		extra := frac[2:]
		if strings.TrimRight(extra, "0") != "" {
			return 0, fmt.Errorf("money has more than 2 decimals: %q", raw)
		}
		frac = frac[:2]
	}
	for len(frac) < 2 {
		frac += "0"
	}
	var f int64
	if frac != "" {
		v, err := strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid money frac part: %w", err)
		}
		f = v
	}
	return i*100 + f, nil
}

func formatCents(amountCents int64) string {
	if amountCents < 0 {
		amountCents = -amountCents
	}
	return fmt.Sprintf("%d.%02d", amountCents/100, amountCents%100)
}

func buildNotifyURL(publicBaseURL string, provider string) (string, error) {
	base := strings.TrimSpace(publicBaseURL)
	if base == "" {
		return "", ErrPaymentPublicBaseURLRequired
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", ErrPaymentPublicBaseURLRequired
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/api/v1/payments/notify/" + provider
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func buildReturnURL(publicBaseURL string) string {
	base := strings.TrimSpace(publicBaseURL)
	if base == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/purchase"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func mergeQueryLikeParams(dst map[string]string, src map[string][]string) {
	if dst == nil || src == nil {
		return
	}
	for k, vs := range src {
		if len(vs) == 0 {
			continue
		}
		dst[k] = vs[0]
	}
}

func parseNotifyParams(rawBody []byte, query map[string][]string) map[string]string {
	out := make(map[string]string)
	mergeQueryLikeParams(out, query)

	body := bytes.TrimSpace(rawBody)
	if len(body) == 0 {
		return out
	}

	// Try form-urlencoded first.
	if v, err := url.ParseQuery(string(body)); err == nil && len(v) > 0 {
		for k := range v {
			out[k] = v.Get(k)
		}
		return out
	}

	// Try JSON object.
	if len(body) > 0 && (body[0] == '{' || body[0] == '[') {
		var obj any
		if err := json.Unmarshal(body, &obj); err == nil {
			switch vv := obj.(type) {
			case map[string]any:
				for k, val := range vv {
					out[k] = fmt.Sprint(val)
				}
			}
		}
	}

	return out
}

func parsePaidFlag(params map[string]string) (bool, bool) {
	// returns (paid, ok)
	if v := strings.TrimSpace(params["trade_status"]); v != "" {
		up := strings.ToUpper(v)
		if up == "TRADE_SUCCESS" || up == "TRADE_FINISHED" || strings.Contains(up, "SUCCESS") {
			return true, true
		}
		if up == "WAIT_BUYER_PAY" || strings.Contains(up, "WAIT") {
			return false, true
		}
		// unknown value, keep searching
	}

	if v := strings.TrimSpace(params["status"]); v != "" {
		switch strings.ToLower(v) {
		case "1", "paid", "success", "succeeded", "true", "ok":
			return true, true
		case "0", "unpaid", "pending", "false", "fail", "failed":
			return false, true
		}
	}

	if v := strings.TrimSpace(params["pay_status"]); v != "" {
		switch strings.ToLower(v) {
		case "success", "paid", "true", "1":
			return true, true
		case "pending", "unpaid", "false", "0":
			return false, true
		}
	}

	return false, false
}

func nowForNotify(ctx context.Context) time.Time {
	_ = ctx
	return time.Now()
}
