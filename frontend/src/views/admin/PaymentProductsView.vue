<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadProducts"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button @click="openCreateDialog" class="btn btn-primary">
            <Icon name="plus" size="md" class="mr-1" />
            {{ t('admin.payments.products.create') }}
          </button>
        </div>
      </template>

      <template #filters>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="max-w-md flex-1">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.payments.products.searchPlaceholder')"
              class="input"
              @input="handleSearch"
            />
          </div>
          <div class="flex gap-2">
            <Select v-model="filters.kind" :options="kindOptions" class="w-40" @change="loadProducts" />
            <Select v-model="filters.status" :options="statusOptions" class="w-36" @change="loadProducts" />
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="products" :loading="loading">
          <template #cell-kind="{ value }">
            <span class="badge badge-gray">
              {{ value === 'subscription' ? t('admin.payments.products.kindSubscription') : t('admin.payments.products.kindBalance') }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', value === 'active' ? 'badge-success' : 'badge-gray']">
              {{ value === 'active' ? t('admin.payments.products.statusActive') : t('admin.payments.products.statusInactive') }}
            </span>
          </template>

          <template #cell-price="{ row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatMoney(row.currency, row.price_cents) }}
            </span>
          </template>

          <template #cell-subscription="{ row }">
            <div class="text-sm text-gray-600 dark:text-gray-300">
              <div v-if="row.kind === 'subscription'">
                <div class="truncate">
                  {{ t('admin.payments.products.fields.group') }}:
                  <span class="font-mono">{{ row.group_id ?? '-' }}</span>
                  <span v-if="row.group_id && groupNameById[row.group_id]" class="ml-1 text-gray-500 dark:text-dark-400">
                    ({{ groupNameById[row.group_id] }})
                  </span>
                </div>
                <div>
                  {{ t('admin.payments.products.fields.validityDays') }}:
                  <span class="font-mono">{{ row.validity_days ?? '-' }}</span>
                </div>
              </div>
              <div v-else class="text-gray-500 dark:text-dark-400">-</div>
            </div>
          </template>

          <template #cell-updated_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-1">
          <button
            @click="openEditDialog(row)"
            class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
            :title="t('common.edit')"
          >
            <Icon name="edit" size="sm" />
          </button>
          <button
            @click="toggleProductStatus(row)"
            class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
            :title="row.status === 'active' ? t('admin.payments.products.actions.deactivate') : t('admin.payments.products.actions.activate')"
            :disabled="togglingId === row.id"
          >
            <Icon :name="row.status === 'active' ? 'ban' : 'checkCircle'" size="sm" />
          </button>
          <button
            @click="openDeleteDialog(row)"
            class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-red-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-300"
            :title="t('common.delete')"
            :disabled="deletingId === row.id"
          >
            <Icon name="trash" size="sm" />
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

    <!-- Create/Edit Dialog -->
    <BaseDialog
      :show="showDialog"
      :title="isEditing ? t('admin.payments.products.edit') : t('admin.payments.products.create')"
      width="normal"
      @close="closeDialog"
    >
      <form id="payment-product-form" @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.payments.products.fields.kind') }}</label>
          <Select v-model="form.kind" :options="kindOptionsForForm" />
        </div>

        <div>
          <label class="input-label">{{ t('admin.payments.products.fields.name') }}</label>
          <input v-model="form.name" type="text" class="input" required />
        </div>

        <div>
          <label class="input-label">{{ t('admin.payments.products.fields.description') }}</label>
          <textarea v-model="form.description_md" rows="3" class="input"></textarea>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.payments.products.fields.status') }}</label>
            <Select v-model="form.status" :options="statusOptionsForForm" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payments.products.fields.sortOrder') }}</label>
            <input v-model.number="form.sort_order" type="number" class="input" />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.payments.products.fields.currency') }}</label>
            <Select v-model="form.currency" :options="currencyOptions" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.payments.products.fields.price') }}</label>
            <input v-model.number="form.price_amount" type="number" min="0" step="0.01" class="input" />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.payments.products.fields.priceHint') }}
            </p>
          </div>
        </div>

        <div v-if="form.kind === 'subscription'" class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
          <p class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('admin.payments.products.subscriptionConfig') }}
          </p>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('admin.payments.products.fields.group') }}</label>
              <Select v-model="form.group_id" :options="subscriptionGroupOptions" />
              <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.payments.products.fields.groupHint') }}
              </p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.payments.products.fields.validityDays') }}</label>
              <input v-model.number="form.validity_days" type="number" min="1" class="input" />
            </div>
          </div>
        </div>

        <div v-else class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
          <p class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('admin.payments.products.balanceConfig') }}
          </p>
          <div class="flex items-center justify-between">
            <div>
              <label class="font-medium text-gray-900 dark:text-white">
                {{ t('admin.payments.products.fields.allowCustomAmount') }}
              </label>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.payments.products.fields.allowCustomAmountHint') }}</p>
            </div>
            <Toggle v-model="form.allow_custom_amount" />
          </div>

          <div v-if="form.allow_custom_amount" class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('admin.payments.products.fields.minAmount') }}</label>
              <input v-model.number="form.min_amount" type="number" min="0" step="0.01" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.payments.products.fields.maxAmount') }}</label>
              <input v-model.number="form.max_amount" type="number" min="0" step="0.01" class="input" />
            </div>
            <div class="md:col-span-2">
              <label class="input-label">{{ t('admin.payments.products.fields.suggestedAmounts') }}</label>
              <input v-model="form.suggested_amounts" type="text" class="input font-mono text-sm" :placeholder="t('admin.payments.products.fields.suggestedAmountsPlaceholder')" />
              <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.payments.products.fields.suggestedAmountsHint') }}
              </p>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('admin.payments.products.fields.exchangeRate') }}</label>
              <input v-model.number="form.exchange_rate" type="number" min="0" step="0.00000001" class="input" />
              <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.payments.products.fields.exchangeRateHint') }}
              </p>
            </div>
            <div>
              <label class="input-label">
                {{ t('admin.payments.products.fields.creditBalance') }}
                <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
              </label>
              <input v-model.number="form.credit_balance" type="number" min="0" step="0.00000001" class="input" />
              <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.payments.products.fields.creditBalanceHint') }}
              </p>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeDialog">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="payment-product-form" :disabled="saving" class="btn btn-primary">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.payments.products.actions.deleteTitle')"
      :message="t('admin.payments.products.actions.deleteConfirm', { name: deletingProduct?.name || '' })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
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
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import type { AdminGroup, PaymentProduct, CreatePaymentProductRequest, UpdatePaymentProductRequest } from '@/types'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

// State
const loading = ref(false)
const saving = ref(false)
const products = ref<PaymentProduct[]>([])
const searchQuery = ref('')

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  kind: '',
  status: ''
})

const groupNameById = reactive<Record<number, string>>({})
const subscriptionGroups = ref<AdminGroup[]>([])

// Dialog
const showDialog = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const togglingId = ref<number | null>(null)
const showDeleteDialog = ref(false)
const deletingProduct = ref<PaymentProduct | null>(null)
const deletingId = ref<number | null>(null)

const resetForm = () => {
  form.id = 0
  form.kind = 'subscription'
  form.name = ''
  form.description_md = ''
  form.status = 'inactive'
  form.sort_order = 0
  form.currency = 'CNY'
  form.price_amount = 0
  form.group_id = 0
  form.validity_days = 30
  form.allow_custom_amount = false
  form.min_amount = 0
  form.max_amount = 0
  form.suggested_amounts = ''
  form.exchange_rate = 1
  form.credit_balance = 0
}

const form = reactive({
  id: 0,
  kind: 'subscription',
  name: '',
  description_md: '',
  status: 'inactive',
  sort_order: 0,
  currency: 'CNY',
  price_amount: 0,
  group_id: 0,
  validity_days: 30,
  allow_custom_amount: false,
  min_amount: 0,
  max_amount: 0,
  suggested_amounts: '',
  exchange_rate: 1,
  credit_balance: 0
})

// Options
const kindOptions = computed(() => [
  { value: '', label: t('admin.payments.products.allKinds') },
  { value: 'subscription', label: t('admin.payments.products.kindSubscription') },
  { value: 'balance', label: t('admin.payments.products.kindBalance') }
])

const kindOptionsForForm = computed(() => [
  { value: 'subscription', label: t('admin.payments.products.kindSubscription') },
  { value: 'balance', label: t('admin.payments.products.kindBalance') }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.payments.products.allStatus') },
  { value: 'active', label: t('admin.payments.products.statusActive') },
  { value: 'inactive', label: t('admin.payments.products.statusInactive') }
])

const statusOptionsForForm = computed(() => [
  { value: 'active', label: t('admin.payments.products.statusActive') },
  { value: 'inactive', label: t('admin.payments.products.statusInactive') }
])

const currencyOptions = computed(() => [
  { value: 'CNY', label: 'CNY' },
  { value: 'USD', label: 'USD' }
])

const subscriptionGroupOptions = computed(() => {
  const options = subscriptionGroups.value.map((g) => ({
    value: g.id,
    label: `${g.id} - ${g.name}${g.status !== 'active' ? ` (${g.status})` : ''}`
  }))
  return [{ value: 0, label: t('admin.payments.products.fields.groupSelectPlaceholder') }, ...options]
})

const columns = computed<Column[]>(() => [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'kind', label: t('admin.payments.products.columns.kind'), sortable: true },
  { key: 'name', label: t('admin.payments.products.columns.name') },
  { key: 'status', label: t('admin.payments.products.columns.status'), sortable: true },
  { key: 'price', label: t('admin.payments.products.columns.price') },
  { key: 'subscription', label: t('admin.payments.products.columns.subscription') },
  { key: 'sort_order', label: t('admin.payments.products.columns.sortOrder'), sortable: true },
  { key: 'updated_at', label: t('admin.payments.products.columns.updatedAt'), sortable: true },
  { key: 'actions', label: t('admin.payments.products.columns.actions') }
])

// Helpers
const toCents = (amount: number) => {
  const n = Number.isFinite(amount) ? amount : 0
  return Math.round(n * 100)
}

const parseSuggestedAmounts = (input: string) => {
  const raw = (input || '').trim()
  if (!raw) return []
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .map((s) => toCents(Number(s)))
    .filter((v) => v > 0)
}

const formatMoney = (currency: string, cents: number) => {
  const c = (currency || 'CNY').toUpperCase()
  const amount = (cents || 0) / 100
  const formatted = amount.toFixed(2)
  if (c === 'CNY') return `¥${formatted}`
  if (c === 'USD') return `$${formatted}`
  return `${formatted} ${c}`
}

// API
let abortController: AbortController | null = null

const loadSubscriptionGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    subscriptionGroups.value = (groups || []).filter((g) => g.subscription_type === 'subscription')
    Object.keys(groupNameById).forEach((k) => delete groupNameById[Number(k)])
    for (const g of groups || []) {
      groupNameById[g.id] = g.name
    }
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const loadProducts = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true

  try {
    const resp = await adminAPI.payments.listProducts(pagination.page, pagination.page_size, {
      kind: filters.kind || undefined,
      status: filters.status || undefined,
      search: searchQuery.value || undefined
    })
    if (currentController.signal.aborted) return

    products.value = resp.items
    pagination.total = resp.total
  } catch (error: any) {
    if (currentController.signal.aborted || error?.name === 'AbortError') return
    appStore.showError(t('admin.payments.products.failedToLoad'))
    console.error('Error loading payment products:', error)
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
    loadProducts()
  }, 300)
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadProducts()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadProducts()
}

const openCreateDialog = () => {
  isEditing.value = false
  editingId.value = null
  resetForm()
  showDialog.value = true
}

const openEditDialog = (p: PaymentProduct) => {
  isEditing.value = true
  editingId.value = p.id
  resetForm()

  form.id = p.id
  form.kind = p.kind
  form.name = p.name
  form.description_md = p.description_md || ''
  form.status = p.status
  form.sort_order = p.sort_order
  form.currency = p.currency || 'CNY'
  form.price_amount = (p.price_cents || 0) / 100

  form.group_id = p.group_id || 0
  form.validity_days = p.validity_days || 30

  form.allow_custom_amount = !!p.allow_custom_amount
  form.min_amount = (p.min_amount_cents || 0) / 100
  form.max_amount = (p.max_amount_cents || 0) / 100
  form.suggested_amounts = (p.suggested_amounts_cents || []).map((v) => (v / 100).toFixed(2)).join(',')
  form.exchange_rate = p.exchange_rate || 1
  form.credit_balance = p.credit_balance || 0

  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  saving.value = false
}

const buildPayload = (): CreatePaymentProductRequest => {
  const hasExchangeRate = Number.isFinite(form.exchange_rate) && form.exchange_rate > 0
  const hasCreditBalance = !form.allow_custom_amount && Number.isFinite(form.credit_balance) && form.credit_balance > 0

  const payload: CreatePaymentProductRequest = {
    kind: form.kind,
    name: form.name.trim(),
    description_md: (form.description_md || '').trim(),
    status: form.status,
    sort_order: form.sort_order,
    currency: form.currency,
    price_cents: toCents(form.price_amount),
    allow_custom_amount: form.kind === 'balance' ? form.allow_custom_amount : false,
    exchange_rate: form.kind === 'balance' && hasExchangeRate ? form.exchange_rate : undefined,
    credit_balance: form.kind === 'balance' && hasCreditBalance ? form.credit_balance : undefined
  }

  if (form.kind === 'subscription') {
    payload.group_id = form.group_id || 0
    payload.validity_days = form.validity_days || 0
  } else {
    payload.group_id = 0
    payload.validity_days = 0

    if (form.allow_custom_amount) {
      payload.price_cents = 0
      payload.min_amount_cents = toCents(form.min_amount)
      payload.max_amount_cents = toCents(form.max_amount)
      payload.suggested_amounts_cents = parseSuggestedAmounts(form.suggested_amounts)
    } else {
      payload.min_amount_cents = 0
      payload.max_amount_cents = 0
      payload.suggested_amounts_cents = []
    }
  }

  return payload
}

const toggleProductStatus = async (p: PaymentProduct) => {
  if (!p?.id) return
  const nextStatus = p.status === 'active' ? 'inactive' : 'active'
  togglingId.value = p.id
  try {
    await adminAPI.payments.updateProduct(p.id, { status: nextStatus })
    appStore.showSuccess(t('admin.payments.products.actions.statusUpdated'))
    loadProducts()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.payments.products.saveFailed'))
    console.error('Toggle payment product status error:', error)
  } finally {
    togglingId.value = null
  }
}

const openDeleteDialog = (p: PaymentProduct) => {
  deletingProduct.value = p
  showDeleteDialog.value = true
}

const cancelDelete = () => {
  showDeleteDialog.value = false
  deletingProduct.value = null
}

const confirmDelete = async () => {
  const p = deletingProduct.value
  if (!p?.id) return
  deletingId.value = p.id
  try {
    await adminAPI.payments.deleteProduct(p.id)
    appStore.showSuccess(t('admin.payments.products.actions.deleted'))
    cancelDelete()
    loadProducts()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.payments.products.saveFailed'))
    console.error('Delete payment product error:', error)
  } finally {
    deletingId.value = null
  }
}

const handleSubmit = async () => {
  saving.value = true
  try {
    const payload = buildPayload()
    if (!isEditing.value) {
      await adminAPI.payments.createProduct(payload)
      appStore.showSuccess(t('admin.payments.products.created'))
    } else {
      const id = editingId.value
      if (!id) return
      const updatePayload: UpdatePaymentProductRequest = payload
      await adminAPI.payments.updateProduct(id, updatePayload)
      appStore.showSuccess(t('admin.payments.products.updated'))
    }
    closeDialog()
    loadProducts()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.payments.products.saveFailed'))
    console.error('Save payment product error:', error)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadSubscriptionGroups()
  loadProducts()
})
</script>
