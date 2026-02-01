<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button @click="loadOrders" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #filters>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="max-w-md flex-1">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.payments.orders.searchPlaceholder')"
              class="input"
              @input="handleSearch"
            />
          </div>
          <div class="flex flex-wrap gap-2">
            <Select v-model="filters.kind" :options="kindOptions" class="w-40" @change="loadOrders" />
            <Select v-model="filters.status" :options="statusOptions" class="w-44" @change="loadOrders" />
            <Select v-model="filters.provider" :options="providerOptions" class="w-40" @change="loadOrders" />
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-order_no="{ row }">
            <div class="min-w-0">
              <div class="font-mono text-xs text-gray-900 dark:text-white">{{ row.order_no }}</div>
              <div v-if="row.provider_trade_no" class="mt-1 text-[11px] text-gray-500 dark:text-dark-400">
                {{ t('admin.payments.orders.columns.providerTradeNo') }}:
                <span class="font-mono">{{ row.provider_trade_no }}</span>
              </div>
            </div>
          </template>

          <template #cell-kind="{ value }">
            <span class="badge badge-gray">
              {{ value === 'subscription' ? t('admin.payments.orders.kindSubscription') : value === 'balance' ? t('admin.payments.orders.kindBalance') : value }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', value === 'fulfilled' ? 'badge-success' : value === 'paid' ? 'badge-primary' : value === 'failed' ? 'badge-danger' : 'badge-gray']">
              {{ formatStatusLabel(value) }}
            </span>
          </template>

          <template #cell-amount_cents="{ row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatMoney(row.currency, row.amount_cents) }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-paid_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ value ? formatDateTime(value) : '-' }}</span>
          </template>

          <template #cell-fulfilled_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ value ? formatDateTime(value) : '-' }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-1">
              <button
                @click="openDetail(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                :title="t('common.view')"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                v-if="row.pay_url"
                @click="openPayUrl(row.pay_url)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                :title="t('admin.payments.orders.actions.openPayUrl')"
              >
                <Icon name="externalLink" size="sm" />
              </button>
              <button
                @click="copyText(row.order_no)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                :title="t('admin.payments.orders.actions.copyOrderNo')"
              >
                <Icon name="copy" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showDetailDialog"
      :title="t('admin.payments.orders.detailTitle')"
      width="normal"
      @close="closeDetail"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.orderNo') }}</div>
              <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">{{ selectedOrder.order_no }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.userId') }}</div>
              <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">{{ selectedOrder.user_id }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.kind') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatKindLabel(selectedOrder.kind) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.provider') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedOrder.provider }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.amount') }}</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatMoney(selectedOrder.currency, selectedOrder.amount_cents) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.status') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatStatusLabel(selectedOrder.status) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.createdAt') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatDateTime(selectedOrder.created_at) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.paidAt') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedOrder.paid_at ? formatDateTime(selectedOrder.paid_at) : '-' }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.fulfilledAt') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedOrder.fulfilled_at ? formatDateTime(selectedOrder.fulfilled_at) : '-' }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.orders.fields.productId') }}</div>
              <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">{{ selectedOrder.product_id ?? '-' }}</div>
            </div>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <button class="btn btn-secondary btn-sm" @click="copyText(selectedOrder.order_no)">
              <Icon name="copy" size="sm" class="mr-1.5" />
              {{ t('admin.payments.orders.actions.copyOrderNo') }}
            </button>
            <button v-if="selectedOrder.pay_url" class="btn btn-secondary btn-sm" @click="openPayUrl(selectedOrder.pay_url)">
              <Icon name="externalLink" size="sm" class="mr-1.5" />
              {{ t('admin.payments.orders.actions.openPayUrl') }}
            </button>
            <button v-if="selectedOrder.pay_url" class="btn btn-secondary btn-sm" @click="copyText(selectedOrder.pay_url)">
              <Icon name="copy" size="sm" class="mr-1.5" />
              {{ t('admin.payments.orders.actions.copyPayUrl') }}
            </button>
          </div>
        </div>

        <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
          <div class="mb-3 flex items-center justify-between gap-3">
            <div class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.payments.orders.notificationsTitle') }}
            </div>
            <button class="btn btn-secondary btn-sm" :disabled="loadingNotifications" @click="loadNotifications">
              <Icon name="refresh" size="sm" class="mr-1.5" :class="loadingNotifications ? 'animate-spin' : ''" />
              {{ t('common.refresh') }}
            </button>
          </div>

          <div v-if="loadingNotifications" class="flex items-center justify-center py-6">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <div v-else-if="notifications.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.payments.orders.noNotifications') }}
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="n in notifications"
              :key="n.id"
              class="rounded-lg border border-gray-100 p-3 dark:border-dark-700"
            >
              <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                <div class="min-w-0">
                  <div class="text-xs text-gray-500 dark:text-dark-400">
                    <span class="mr-2 font-mono">{{ n.provider }}</span>
                    <span class="font-mono">{{ n.event_id }}</span>
                  </div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                    {{ formatDateTime(n.received_at) }}
                  </div>
                </div>
                <div class="flex flex-wrap gap-2">
                  <span :class="['badge', n.verified ? 'badge-success' : 'badge-gray']">
                    {{ n.verified ? t('admin.payments.orders.verified') : t('admin.payments.orders.notVerified') }}
                  </span>
                  <span :class="['badge', n.processed ? 'badge-success' : 'badge-gray']">
                    {{ n.processed ? t('admin.payments.orders.processed') : t('admin.payments.orders.notProcessed') }}
                  </span>
                </div>
              </div>

              <div v-if="n.process_error" class="mt-2 text-xs text-red-700 dark:text-red-300">
                {{ n.process_error }}
              </div>

              <div class="mt-2 flex flex-wrap gap-2">
                <button class="btn btn-secondary btn-xs" @click="copyText(n.raw_body)">
                  <Icon name="copy" size="xs" class="mr-1.5" />
                  {{ t('admin.payments.orders.actions.copyRaw') }}
                </button>
                <button class="btn btn-secondary btn-xs" @click="toggleRaw(n.id)">
                  <Icon :name="expandedRaw[n.id] ? 'eyeOff' : 'eye'" size="xs" class="mr-1.5" />
                  {{ expandedRaw[n.id] ? t('admin.payments.orders.actions.hideRaw') : t('admin.payments.orders.actions.showRaw') }}
                </button>
              </div>

              <pre
                v-if="expandedRaw[n.id]"
                class="mt-2 max-h-64 overflow-auto rounded-lg bg-gray-50 p-3 text-xs text-gray-800 dark:bg-dark-800 dark:text-dark-100"
              ><code>{{ n.raw_body }}</code></pre>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <button type="button" class="btn btn-secondary" @click="closeDetail">
            {{ t('common.close') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import type { PaymentOrder, PaymentNotification } from '@/types'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const orders = ref<PaymentOrder[]>([])
const searchQuery = ref('')

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  kind: '',
  status: '',
  provider: ''
})

const kindOptions = computed(() => [
  { value: '', label: t('admin.payments.orders.allKinds') },
  { value: 'subscription', label: t('admin.payments.orders.kindSubscription') },
  { value: 'balance', label: t('admin.payments.orders.kindBalance') }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.payments.orders.allStatus') },
  { value: 'created', label: t('admin.payments.orders.statusCreated') },
  { value: 'paid', label: t('admin.payments.orders.statusPaid') },
  { value: 'fulfilled', label: t('admin.payments.orders.statusFulfilled') },
  { value: 'cancelled', label: t('admin.payments.orders.statusCancelled') },
  { value: 'expired', label: t('admin.payments.orders.statusExpired') },
  { value: 'failed', label: t('admin.payments.orders.statusFailed') }
])

const providerOptions = computed(() => [
  { value: '', label: t('admin.payments.orders.allProviders') },
  { value: 'epay', label: 'Epay' },
  { value: 'tokenpay', label: 'TokenPay' },
  { value: 'manual', label: 'Manual' }
])

const columns = computed<Column[]>(() => [
  { key: 'order_no', label: t('admin.payments.orders.columns.orderNo') },
  { key: 'user_id', label: t('admin.payments.orders.columns.userId'), sortable: true },
  { key: 'kind', label: t('admin.payments.orders.columns.kind'), sortable: true },
  { key: 'provider', label: t('admin.payments.orders.columns.provider'), sortable: true },
  { key: 'amount_cents', label: t('admin.payments.orders.columns.amount') },
  { key: 'status', label: t('admin.payments.orders.columns.status'), sortable: true },
  { key: 'created_at', label: t('admin.payments.orders.columns.createdAt'), sortable: true },
  { key: 'paid_at', label: t('admin.payments.orders.columns.paidAt'), sortable: true },
  { key: 'fulfilled_at', label: t('admin.payments.orders.columns.fulfilledAt'), sortable: true },
  { key: 'actions', label: t('admin.payments.orders.columns.actions') }
])

const formatMoney = (currency: string, cents: number) => {
  const c = (currency || 'CNY').toUpperCase()
  const amount = (cents || 0) / 100
  const formatted = amount.toFixed(2)
  if (c === 'CNY') return `¥${formatted}`
  if (c === 'USD') return `$${formatted}`
  return `${formatted} ${c}`
}

const formatKindLabel = (kind: string) => {
  if (kind === 'subscription') return t('admin.payments.orders.kindSubscription')
  if (kind === 'balance') return t('admin.payments.orders.kindBalance')
  return kind
}

const formatStatusLabel = (status: string) => {
  switch (status) {
    case 'created':
      return t('admin.payments.orders.statusCreated')
    case 'paid':
      return t('admin.payments.orders.statusPaid')
    case 'fulfilled':
      return t('admin.payments.orders.statusFulfilled')
    case 'cancelled':
      return t('admin.payments.orders.statusCancelled')
    case 'expired':
      return t('admin.payments.orders.statusExpired')
    case 'failed':
      return t('admin.payments.orders.statusFailed')
    default:
      return status
  }
}

const openPayUrl = (url: string | null) => {
  const u = (url || '').trim()
  if (!u) return
  window.open(u, '_blank', 'noopener,noreferrer')
}

const copyText = async (text: string | null) => {
  const v = (text || '').trim()
  if (!v) return
  try {
    await navigator.clipboard.writeText(v)
    appStore.showSuccess(t('common.copied'))
  } catch {
    // ignore
  }
}

let abortController: AbortController | null = null

const loadOrders = async () => {
  if (abortController) abortController.abort()
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true

  try {
    const resp = await adminAPI.payments.listOrders(pagination.page, pagination.page_size, {
      kind: filters.kind || undefined,
      status: filters.status || undefined,
      provider: filters.provider || undefined,
      search: searchQuery.value || undefined
    }, { signal: currentController.signal })
    if (currentController.signal.aborted) return
    orders.value = resp.items
    pagination.total = resp.total
  } catch (error: any) {
    if (currentController.signal.aborted || error?.name === 'AbortError') return
    appStore.showError(t('admin.payments.orders.failedToLoad'))
    console.error('Error loading payment orders:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
      loading.value = false
      abortController = null
    }
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadOrders()
  }, 300)
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadOrders()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadOrders()
}

// Detail dialog + notifications
const showDetailDialog = ref(false)
const selectedOrder = ref<PaymentOrder | null>(null)
const notifications = ref<PaymentNotification[]>([])
const loadingNotifications = ref(false)
const expandedRaw = ref<Record<number, boolean>>({})

const openDetail = async (o: PaymentOrder) => {
  selectedOrder.value = o
  showDetailDialog.value = true
  expandedRaw.value = {}
  await loadNotifications()
}

const closeDetail = () => {
  showDetailDialog.value = false
  selectedOrder.value = null
  notifications.value = []
  loadingNotifications.value = false
  expandedRaw.value = {}
}

const toggleRaw = (id: number) => {
  expandedRaw.value[id] = !expandedRaw.value[id]
}

const loadNotifications = async () => {
  const o = selectedOrder.value
  if (!o?.order_no) return
  loadingNotifications.value = true
  try {
    const resp = await adminAPI.payments.listNotifications(1, 20, { order_no: o.order_no })
    notifications.value = resp.items || []
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.payments.orders.failedToLoadNotifications'))
    console.error('Failed to load payment notifications:', error)
  } finally {
    loadingNotifications.value = false
  }
}

onMounted(() => {
  loadOrders()
})
</script>
