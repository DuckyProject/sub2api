<template>
  <AppLayout>
    <div class="purchase-page-layout">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('purchase.title') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ purchaseMode === 'native' ? t('purchase.nativeDescription') : t('purchase.description') }}
          </p>
        </div>

        <div class="flex items-center gap-2">
          <a
            v-if="purchaseMode === 'iframe' && isValidUrl"
            :href="purchaseUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
            {{ t('purchase.openInNewTab') }}
          </a>
        </div>
      </div>

      <div class="card flex-1 min-h-0 overflow-hidden">
        <div v-if="loading" class="flex h-full items-center justify-center py-12">
          <div
            class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
          ></div>
        </div>

        <div
          v-else-if="!purchaseEnabled"
          class="flex h-full items-center justify-center p-10 text-center"
        >
          <div class="max-w-md">
            <div
              class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
            >
              <Icon name="creditCard" size="lg" class="text-gray-400" />
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('purchase.notEnabledTitle') }}
            </h3>
            <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
              {{ t('purchase.notEnabledDesc') }}
            </p>
          </div>
        </div>

        <div
          v-else-if="!isValidUrl"
          class="flex h-full items-center justify-center p-10 text-center"
        >
          <div class="max-w-md">
            <div
              class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
            >
              <Icon name="link" size="lg" class="text-gray-400" />
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('purchase.notConfiguredTitle') }}
            </h3>
            <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
              {{ t('purchase.notConfiguredDesc') }}
            </p>
          </div>
        </div>

        <div v-else-if="purchaseMode === 'native'" class="h-full overflow-y-auto p-6">
          <div class="space-y-6">
            <div
              v-if="!paymentEnabled"
              class="card border-amber-200 bg-amber-50 dark:border-amber-800/50 dark:bg-amber-900/20"
            >
              <div class="p-6">
                <div class="flex items-start gap-4">
                  <div
                    class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 dark:bg-amber-900/30"
                  >
                    <Icon name="exclamationCircle" size="md" class="text-amber-700 dark:text-amber-300" />
                  </div>
                  <div class="flex-1">
                    <h3 class="text-sm font-semibold text-amber-800 dark:text-amber-200">
                      {{ t('purchase.paymentNotEnabledTitle') }}
                    </h3>
                    <p class="mt-2 text-sm text-amber-700 dark:text-amber-300">
                      {{ t('purchase.paymentNotEnabledDesc') }}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="space-y-6">
              <!-- Payment method -->
              <div class="card">
                <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                    {{ t('purchase.selectPaymentMethod') }}
                  </h3>
                </div>
                <div class="p-6">
                  <div class="flex flex-wrap gap-2">
                    <button
                      v-for="m in paymentMethods"
                      :key="m"
                      type="button"
                      class="btn btn-sm"
                      :class="selectedProvider === m ? 'btn-primary' : 'btn-secondary'"
                      @click="selectedProvider = m"
                    >
                      {{ m === 'epay' ? 'Epay' : m === 'tokenpay' ? 'TokenPay' : m }}
                    </button>
                  </div>
                  <p class="mt-3 text-xs text-gray-500 dark:text-dark-400">
                    {{ t('purchase.paidTip') }}
                  </p>
                </div>
              </div>

              <!-- Products -->
              <div class="card">
                <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
                  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                      {{ activeKind === 'subscription' ? t('purchase.packagesTitle') : t('purchase.balanceTitle') }}
                    </h3>
                    <div class="flex gap-2">
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="activeKind === 'subscription' ? 'btn-primary' : 'btn-secondary'"
                        @click="switchKind('subscription')"
                      >
                        {{ t('purchase.tabs.subscription') }}
                      </button>
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="activeKind === 'balance' ? 'btn-primary' : 'btn-secondary'"
                        @click="switchKind('balance')"
                      >
                        {{ t('purchase.tabs.balance') }}
                      </button>
                    </div>
                  </div>
                </div>
                <div class="p-6">
                  <div v-if="loadingProducts" class="flex items-center justify-center py-10">
                    <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
                  </div>

                  <div v-else-if="productsError" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800/50 dark:bg-red-900/20 dark:text-red-300">
                    {{ productsError }}
                  </div>

                  <div v-else-if="products.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
                    {{ t('purchase.noProducts') }}
                  </div>

                  <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                    <div
                      v-for="p in products"
                      :key="p.id"
                      class="rounded-xl border border-gray-100 p-5 dark:border-dark-700"
                    >
                      <div class="flex items-start justify-between gap-3">
                        <div class="min-w-0">
                          <h4 class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                            {{ p.name }}
                          </h4>
                          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                            {{ p.kind === 'subscription' ? t('purchase.tabs.subscription') : p.kind === 'balance' ? t('purchase.tabs.balance') : p.kind }}
                          </p>
                        </div>
                        <div class="flex-shrink-0 text-right">
                          <div class="text-sm font-semibold text-gray-900 dark:text-white">
                            <span v-if="p.kind === 'balance' && p.allow_custom_amount">
                              {{ t('purchase.customAmountTag') }}
                            </span>
                            <span v-else>
                              {{ formatAmount(p.currency, p.price_cents) }}
                            </span>
                          </div>
                        </div>
                      </div>

                      <div v-if="activeKind === 'balance'" class="mt-2 text-xs text-gray-500 dark:text-dark-400">
                        {{ formatBalanceCreditPreview(p) }}
                      </div>

                      <p
                        v-if="p.description_md"
                        class="mt-3 whitespace-pre-line text-sm text-gray-600 dark:text-dark-300"
                        v-text="p.description_md"
                      ></p>

                      <div v-if="activeKind === 'balance' && p.allow_custom_amount" class="mt-4 space-y-2">
                        <label class="block text-xs font-medium text-gray-700 dark:text-dark-200">
                          {{ t('purchase.customAmountLabel') }}
                        </label>
                        <div class="flex items-center gap-2">
                          <input
                            v-model="customAmountByProduct[p.id]"
                            type="text"
                            class="input"
                            :placeholder="t('purchase.customAmountPlaceholder')"
                            @input="clearAmountError(p.id)"
                          />
                        </div>
                        <p class="text-xs text-gray-500 dark:text-dark-400">
                          {{ formatMinMaxHint(p) }}
                        </p>

                        <div v-if="(p.suggested_amounts_cents || []).length > 0" class="flex flex-wrap gap-2">
                          <button
                            v-for="amt in p.suggested_amounts_cents"
                            :key="amt"
                            type="button"
                            class="btn btn-secondary btn-xs"
                            @click="setSuggestedAmount(p.id, amt)"
                          >
                            {{ formatAmount(p.currency, amt) }}
                          </button>
                        </div>

                        <p v-if="amountErrorByProduct[p.id]" class="text-xs text-red-700 dark:text-red-300">
                          {{ amountErrorByProduct[p.id] }}
                        </p>
                      </div>

                      <button
                        type="button"
                        class="btn btn-primary mt-4 w-full"
                        :disabled="creatingOrder || !selectedProvider"
                        @click="createOrder(p)"
                      >
                        <span v-if="creatingOrder && creatingProductId === p.id" class="inline-flex items-center">
                          <span class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/80 border-t-transparent"></span>
                          {{ t('purchase.creatingOrder') }}
                        </span>
                        <span v-else>{{ activeKind === 'balance' ? t('purchase.topupNow') : t('purchase.buyNow') }}</span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Order -->
              <transition name="fade">
                <div v-if="createdOrder" class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20">
                  <div class="p-6">
                    <div class="flex items-start justify-between gap-4">
                      <div class="min-w-0">
                        <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-200">
                          {{ t('purchase.orderCreatedTitle') }}
                        </h3>
                        <div class="mt-2 space-y-1 text-sm text-emerald-800/90 dark:text-emerald-200/90">
                          <p><span class="font-medium">{{ t('purchase.orderNo') }}:</span> <span class="font-mono">{{ createdOrder.order_no }}</span></p>
                          <p><span class="font-medium">{{ t('purchase.amount') }}:</span> {{ formatAmount(createdOrder.currency, createdOrder.amount_cents) }}</p>
                          <p><span class="font-medium">{{ t('purchase.status') }}:</span> {{ createdOrder.status }}</p>
                        </div>
                      </div>
                      <div class="flex flex-col gap-2">
                        <button type="button" class="btn btn-secondary btn-sm" :disabled="!createdOrder.pay_url" @click="openPayUrl(createdOrder.pay_url)">
                          {{ t('purchase.openPayPage') }}
                        </button>
                        <button type="button" class="btn btn-secondary btn-sm" :disabled="!createdOrder.pay_url" @click="copyPayUrl(createdOrder.pay_url)">
                          {{ copied ? t('purchase.copied') : t('purchase.copyPayLink') }}
                        </button>
                        <button type="button" class="btn btn-secondary btn-sm" :disabled="refreshingOrder" @click="refreshOrderStatus">
                          <span v-if="refreshingOrder" class="inline-flex items-center">
                            <span class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-gray-500/70 border-t-transparent dark:border-dark-300/70"></span>
                            {{ t('purchase.refreshing') }}
                          </span>
                          <span v-else>{{ t('purchase.refreshStatus') }}</span>
                        </button>
                      </div>
                    </div>

                    <p v-if="orderError" class="mt-4 text-sm text-red-700 dark:text-red-300">
                      {{ orderError }}
                    </p>
                  </div>
                </div>
              </transition>
            </div>
          </div>
        </div>

        <iframe v-else :src="purchaseUrl" class="h-full w-full border-0" allowfullscreen></iframe>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { paymentsAPI } from '@/api'
import type { PaymentProduct, PaymentOrder } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(false)

const activeKind = ref<'subscription' | 'balance'>('subscription')
const subscriptionProducts = ref<PaymentProduct[]>([])
const balanceProducts = ref<PaymentProduct[]>([])
const products = computed(() => (activeKind.value === 'subscription' ? subscriptionProducts.value : balanceProducts.value))
const loadingProducts = ref(false)
const productsError = ref('')

const selectedProvider = ref<string>('')
const creatingOrder = ref(false)
const creatingProductId = ref<number | null>(null)
const createdOrder = ref<PaymentOrder | null>(null)
const refreshingOrder = ref(false)
const orderError = ref('')
const copied = ref(false)

const customAmountByProduct = ref<Record<number, string>>({})
const amountErrorByProduct = ref<Record<number, string>>({})

const purchaseEnabled = computed(() => {
  const mode = (appStore.cachedPublicSettings?.purchase_subscription_mode || 'disabled').toLowerCase()
  // 兼容老字段：若后端还没下发 mode，则回退旧的 enabled 逻辑
  if (!appStore.cachedPublicSettings?.purchase_subscription_mode) {
    return appStore.cachedPublicSettings?.purchase_subscription_enabled ?? false
  }
  return mode !== 'disabled'
})

const purchaseUrl = computed(() => {
  const mode = (appStore.cachedPublicSettings?.purchase_subscription_mode || '').toLowerCase()
  if (mode === 'native') return ''
  return (appStore.cachedPublicSettings?.purchase_subscription_url || '').trim()
})

const purchaseMode = computed(() => {
  return (appStore.cachedPublicSettings?.purchase_subscription_mode || 'disabled').toLowerCase()
})

const isValidUrl = computed(() => {
  const url = purchaseUrl.value
  if (purchaseMode.value !== 'iframe') return true
  return url.startsWith('http://') || url.startsWith('https://')
})

const paymentEnabled = computed(() => {
  const s = appStore.cachedPublicSettings
  return !!s?.payment_enabled && Array.isArray(s?.payment_methods) && s!.payment_methods.length > 0
})

const paymentMethods = computed(() => {
  return appStore.cachedPublicSettings?.payment_methods || []
})

const ensureDefaultProvider = () => {
  if (selectedProvider.value) return
  const methods = paymentMethods.value
  if (methods.length > 0) selectedProvider.value = methods[0]
}

const formatAmount = (currency: string, cents: number) => {
  const c = (currency || 'CNY').toUpperCase()
  const amount = (cents || 0) / 100
  const formatted = amount.toFixed(2)
  if (c === 'CNY') return `¥${formatted}`
  if (c === 'USD') return `$${formatted}`
  return `${formatted} ${c}`
}

const formatCentsInput = (cents: number) => {
  const v = Number.isFinite(cents) ? cents : 0
  return (v / 100).toFixed(2)
}

const parseMoneyToCents = (raw: string): number | null => {
  let s = (raw || '').trim()
  if (!s) return null
  if (s.startsWith('+')) s = s.slice(1)
  if (s.startsWith('-')) return null

  const parts = s.split('.')
  if (parts.length > 2) return null
  const intPart = parts[0] === '' ? '0' : parts[0]
  if (!/^[0-9]+$/.test(intPart)) return null

  let frac = parts.length === 2 ? parts[1] : ''
  if (frac && !/^[0-9]+$/.test(frac)) return null
  if (frac.length > 2) {
    const extra = frac.slice(2)
    if (extra.replace(/0/g, '') !== '') return null
    frac = frac.slice(0, 2)
  }
  while (frac.length < 2) frac += '0'

  const i = Number.parseInt(intPart, 10)
  const f = frac ? Number.parseInt(frac, 10) : 0
  if (!Number.isFinite(i) || !Number.isFinite(f)) return null
  return i * 100 + f
}

const clearAmountError = (productId: number) => {
  if (!productId) return
  if (amountErrorByProduct.value[productId]) {
    amountErrorByProduct.value[productId] = ''
  }
}

const formatMinMaxHint = (p: PaymentProduct) => {
  const min = p.min_amount_cents ?? 0
  const max = p.max_amount_cents ?? 0
  if (min > 0 && max > 0) {
    return t('purchase.minMaxHint', {
      min: formatAmount(p.currency, min),
      max: formatAmount(p.currency, max)
    })
  }
  if (min > 0) {
    return t('purchase.minHint', { min: formatAmount(p.currency, min) })
  }
  if (max > 0) {
    return t('purchase.maxHint', { max: formatAmount(p.currency, max) })
  }
  return ''
}

const getCustomAmountForProduct = (p: PaymentProduct): { amount?: string; cents?: number; error?: string } => {
  if (!p.allow_custom_amount) return {}
  const raw = (customAmountByProduct.value[p.id] || '').trim()
  if (!raw) return { error: t('purchase.amountRequired') }
  const cents = parseMoneyToCents(raw)
  if (!cents || cents <= 0) return { error: t('purchase.invalidAmount') }
  if (p.min_amount_cents != null && cents < p.min_amount_cents) {
    return {
      error: t('purchase.amountOutOfRange', {
        min: formatAmount(p.currency, p.min_amount_cents),
        max: p.max_amount_cents != null ? formatAmount(p.currency, p.max_amount_cents) : '-'
      })
    }
  }
  if (p.max_amount_cents != null && cents > p.max_amount_cents) {
    return {
      error: t('purchase.amountOutOfRange', {
        min: p.min_amount_cents != null ? formatAmount(p.currency, p.min_amount_cents) : '-',
        max: formatAmount(p.currency, p.max_amount_cents)
      })
    }
  }
  return { amount: raw, cents }
}

const setSuggestedAmount = (productId: number, cents: number) => {
  if (!productId) return
  customAmountByProduct.value[productId] = formatCentsInput(cents)
  clearAmountError(productId)
}

const formatBalanceCreditPreview = (p: PaymentProduct) => {
  if (p.kind !== 'balance') return ''

  // Fixed credit override (non-custom amount).
  if (!p.allow_custom_amount && p.credit_balance != null && p.credit_balance > 0) {
    const credit = p.credit_balance.toFixed(8).replace(/0+$/, '').replace(/\.$/, '')
    return t('purchase.creditFixedHint', { credit })
  }

  // Rate-based preview when exchange_rate is provided by product.
  const rate = p.exchange_rate != null && p.exchange_rate > 0 ? p.exchange_rate : null
  if (!rate) {
    return t('purchase.creditRateFallbackHint')
  }

  let cents = p.price_cents || 0
  if (p.allow_custom_amount) {
    const r = getCustomAmountForProduct(p)
    if (r.cents != null) cents = r.cents
  }
  const credit = ((cents || 0) / 100) * rate
  return t('purchase.creditRateHint', {
    rate: rate.toFixed(8).replace(/0+$/, '').replace(/\.$/, ''),
    credit: credit.toFixed(8).replace(/0+$/, '').replace(/\.$/, '')
  })
}

const openPayUrl = (url: string | null) => {
  const u = (url || '').trim()
  if (!u) return
  window.open(u, '_blank', 'noopener,noreferrer')
}

const copyPayUrl = async (url: string | null) => {
  copied.value = false
  const u = (url || '').trim()
  if (!u) return
  try {
    await navigator.clipboard.writeText(u)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    // ignore clipboard errors
  }
}

const createOrder = async (p: PaymentProduct) => {
  if (!selectedProvider.value) return
  if (!p?.id) return
  creatingOrder.value = true
  creatingProductId.value = p.id
  orderError.value = ''
  try {
    const clientRequestId = `web-${Date.now()}-${Math.random().toString(16).slice(2)}`
    const payload: any = {
      product_id: p.id,
      provider: selectedProvider.value,
      client_request_id: clientRequestId
    }

    if (p.kind === 'balance' && p.allow_custom_amount) {
      const r = getCustomAmountForProduct(p)
      if (r.error) {
        amountErrorByProduct.value[p.id] = r.error
        return
      }
      payload.amount = r.amount
    }

    const resp = await paymentsAPI.createOrder(payload)
    createdOrder.value = resp.order
  } catch (e: any) {
    orderError.value = e?.message || t('purchase.createOrderFailed')
  } finally {
    creatingOrder.value = false
    creatingProductId.value = null
  }
}

const refreshOrderStatus = async () => {
  if (!createdOrder.value) return
  refreshingOrder.value = true
  orderError.value = ''
  try {
    const updated = await paymentsAPI.getOrder(createdOrder.value.order_no)
    createdOrder.value = updated
    if (updated.status === 'fulfilled' || updated.status === 'paid') {
      // Refresh user profile (balance/subscription may change).
      try {
        await authStore.refreshUser()
      } catch {
        // ignore
      }
    }
  } catch (e: any) {
    orderError.value = e?.message || t('purchase.refreshOrderFailed')
  } finally {
    refreshingOrder.value = false
  }
}

const loadProducts = async (kind: 'subscription' | 'balance') => {
  productsError.value = ''
  loadingProducts.value = true
  try {
    const list = await paymentsAPI.listProducts(kind)
    if (kind === 'subscription') {
      subscriptionProducts.value = list || []
    } else {
      balanceProducts.value = list || []
      // Prefill custom amount inputs for better UX.
      for (const p of list || []) {
        if (p.kind !== 'balance' || !p.allow_custom_amount) continue
        if (customAmountByProduct.value[p.id]) continue
        const suggested = (p.suggested_amounts_cents || [])[0]
        const fallback = suggested || p.min_amount_cents || 0
        if (fallback > 0) {
          customAmountByProduct.value[p.id] = formatCentsInput(fallback)
        }
      }
    }
  } catch (e: any) {
    productsError.value = e?.message || t('purchase.loadProductsFailed')
  } finally {
    loadingProducts.value = false
  }
}

const switchKind = async (kind: 'subscription' | 'balance') => {
  if (activeKind.value === kind) return
  activeKind.value = kind
  if (!paymentEnabled.value) return
  const list = kind === 'subscription' ? subscriptionProducts.value : balanceProducts.value
  if (list.length === 0) {
    await loadProducts(kind)
  }
}

onMounted(async () => {
  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
    ensureDefaultProvider()
    if ((appStore.cachedPublicSettings?.purchase_subscription_mode || '').toLowerCase() === 'native') {
      if (paymentEnabled.value) {
        await loadProducts('subscription')
      }
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.purchase-page-layout {
  @apply flex flex-col gap-6;
  height: calc(100vh - 64px - 4rem); /* 减去 header + lg:p-8 的上下padding */
}
</style>
