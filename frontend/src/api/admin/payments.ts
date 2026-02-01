/**
 * Admin Payments API endpoints
 * Payment products management (subscription packages / balance recharge products)
 */

import { apiClient } from '../client'
import type {
  PaginatedResponse,
  PaymentProduct,
  CreatePaymentProductRequest,
  UpdatePaymentProductRequest,
  PaymentOrder,
  PaymentNotification
} from '@/types'

export async function listProducts(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    kind?: string
    status?: string
    search?: string
  },
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<PaymentProduct>> {
  const { data } = await apiClient.get<PaginatedResponse<PaymentProduct>>('/admin/payments/products', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

export async function createProduct(payload: CreatePaymentProductRequest): Promise<PaymentProduct> {
  const { data } = await apiClient.post<PaymentProduct>('/admin/payments/products', payload)
  return data
}

export async function updateProduct(id: number, payload: UpdatePaymentProductRequest): Promise<PaymentProduct> {
  const { data } = await apiClient.put<PaymentProduct>(`/admin/payments/products/${id}`, payload)
  return data
}

export async function deleteProduct(id: number): Promise<{ deleted: boolean }> {
  const { data } = await apiClient.delete<{ deleted: boolean }>(`/admin/payments/products/${id}`)
  return data
}

export async function listOrders(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    kind?: string
    status?: string
    provider?: string
    user_id?: number
    search?: string
  },
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<PaymentOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<PaymentOrder>>('/admin/payments/orders', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

export async function getOrder(orderNo: string): Promise<PaymentOrder> {
  const { data } = await apiClient.get<PaymentOrder>(`/admin/payments/orders/${encodeURIComponent(orderNo)}`)
  return data
}

export async function listNotifications(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    provider?: string
    order_no?: string
    search?: string
  },
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<PaymentNotification>> {
  const { data } = await apiClient.get<PaginatedResponse<PaymentNotification>>('/admin/payments/notifications', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

export default {
  listProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  listOrders,
  getOrder,
  listNotifications
}
