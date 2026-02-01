/**
 * Payments API
 * Native purchase flow: list products, create orders, query order status.
 */

import apiClient from './client'
import type { PaymentProduct, CreatePaymentOrderRequest, CreatePaymentOrderResponse, PaymentOrder, FetchOptions, BasePaginationResponse } from '@/types'

export const paymentsAPI = {
  async listProducts(kind: string, options: FetchOptions = {}): Promise<PaymentProduct[]> {
    const response = await apiClient.get<PaymentProduct[]>('/payments/products', {
      params: { kind },
      signal: options.signal
    })
    return response.data
  },

  async createOrder(payload: CreatePaymentOrderRequest, options: FetchOptions = {}): Promise<CreatePaymentOrderResponse> {
    const response = await apiClient.post<CreatePaymentOrderResponse>('/payments/orders', payload, {
      signal: options.signal
    })
    return response.data
  },

  async getOrder(orderNo: string, options: FetchOptions = {}): Promise<PaymentOrder> {
    const response = await apiClient.get<PaymentOrder>(`/payments/orders/${encodeURIComponent(orderNo)}`, {
      signal: options.signal
    })
    return response.data
  },

  async listOrders(params: { page?: number; page_size?: number; status?: string } = {}, options: FetchOptions = {}): Promise<BasePaginationResponse<PaymentOrder>> {
    const response = await apiClient.get<BasePaginationResponse<PaymentOrder>>('/payments/orders', {
      params,
      signal: options.signal
    })
    return response.data
  }
}

export default paymentsAPI

