package service

// ProvidePaymentProviders wires all supported payment providers.
//
// Providers should be robust against being disabled/misconfigured at runtime.
// The PaymentService will also gate provider usage via settings and provider map.
func ProvidePaymentProviders(settingService *SettingService) map[string]PaymentProvider {
	return map[string]PaymentProvider{
		PaymentProviderEpay:     &epayProvider{settingService: settingService},
		PaymentProviderTokenPay: &tokenPayProvider{settingService: settingService},
	}
}
