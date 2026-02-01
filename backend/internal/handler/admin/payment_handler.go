package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles admin payment product management.
type PaymentHandler struct {
	paymentAdmin *service.PaymentAdminService
}

func NewPaymentHandler(paymentAdmin *service.PaymentAdminService) *PaymentHandler {
	return &PaymentHandler{paymentAdmin: paymentAdmin}
}

type CreatePaymentProductRequest struct {
	Kind          string `json:"kind" binding:"required,oneof=subscription balance"`
	Name          string `json:"name" binding:"required"`
	DescriptionMD string `json:"description_md"`
	Status        string `json:"status" binding:"omitempty,oneof=active inactive"`
	SortOrder     int    `json:"sort_order"`

	Currency   string `json:"currency"`
	PriceCents int64  `json:"price_cents"`

	GroupID      *int64 `json:"group_id"`
	ValidityDays *int   `json:"validity_days"`

	CreditBalance *float64 `json:"credit_balance"`

	AllowCustomAmount     bool     `json:"allow_custom_amount"`
	MinAmountCents        *int64   `json:"min_amount_cents"`
	MaxAmountCents        *int64   `json:"max_amount_cents"`
	SuggestedAmountsCents []int64  `json:"suggested_amounts_cents"`
	ExchangeRate          *float64 `json:"exchange_rate"`
}

type UpdatePaymentProductRequest struct {
	Kind          *string `json:"kind" binding:"omitempty,oneof=subscription balance"`
	Name          *string `json:"name"`
	DescriptionMD *string `json:"description_md"`
	Status        *string `json:"status" binding:"omitempty,oneof=active inactive"`
	SortOrder     *int    `json:"sort_order"`

	Currency   *string `json:"currency"`
	PriceCents *int64  `json:"price_cents"`

	GroupID      *int64 `json:"group_id"`
	ValidityDays *int   `json:"validity_days"`

	CreditBalance *float64 `json:"credit_balance"`

	AllowCustomAmount     *bool    `json:"allow_custom_amount"`
	MinAmountCents        *int64   `json:"min_amount_cents"`
	MaxAmountCents        *int64   `json:"max_amount_cents"`
	SuggestedAmountsCents *[]int64 `json:"suggested_amounts_cents"`
	ExchangeRate          *float64 `json:"exchange_rate"`
}

// ListOrders
// GET /api/v1/admin/payments/orders?kind=&status=&provider=&search=&user_id=&page=&page_size=
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	kind := strings.TrimSpace(c.Query("kind"))
	status := strings.TrimSpace(c.Query("status"))
	provider := strings.TrimSpace(c.Query("provider"))
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	var userID *int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = &v
	}

	orders, pr, err := h.paymentAdmin.ListOrders(c.Request.Context(), params, userID, kind, status, provider, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentOrder, 0, len(orders))
	for i := range orders {
		o := orders[i]
		out = append(out, dto.PaymentOrder{
			ID:              o.ID,
			OrderNo:         o.OrderNo,
			UserID:          o.UserID,
			Kind:            o.Kind,
			ProductID:       o.ProductID,
			Status:          o.Status,
			Provider:        o.Provider,
			Currency:        o.Currency,
			AmountCents:     o.AmountCents,
			ClientRequestID: o.ClientRequestID,
			ProviderTradeNo: o.ProviderTradeNo,
			PayURL:          o.PayURL,
			ExpiresAt:       o.ExpiresAt,
			PaidAt:          o.PaidAt,
			FulfilledAt:     o.FulfilledAt,
			CreatedAt:       o.CreatedAt,
			UpdatedAt:       o.UpdatedAt,
		})
	}

	if pr == nil {
		response.Paginated(c, out, int64(len(out)), page, pageSize)
		return
	}
	response.Paginated(c, out, pr.Total, pr.Page, pr.PageSize)
}

// GetOrder
// GET /api/v1/admin/payments/orders/:order_no
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		response.BadRequest(c, "Missing order_no")
		return
	}

	o, err := h.paymentAdmin.GetOrder(c.Request.Context(), orderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentOrder{
		ID:              o.ID,
		OrderNo:         o.OrderNo,
		UserID:          o.UserID,
		Kind:            o.Kind,
		ProductID:       o.ProductID,
		Status:          o.Status,
		Provider:        o.Provider,
		Currency:        o.Currency,
		AmountCents:     o.AmountCents,
		ClientRequestID: o.ClientRequestID,
		ProviderTradeNo: o.ProviderTradeNo,
		PayURL:          o.PayURL,
		ExpiresAt:       o.ExpiresAt,
		PaidAt:          o.PaidAt,
		FulfilledAt:     o.FulfilledAt,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	})
}

// ListNotifications
// GET /api/v1/admin/payments/notifications?provider=&order_no=&search=&page=&page_size=
func (h *PaymentHandler) ListNotifications(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	provider := strings.TrimSpace(c.Query("provider"))
	orderNo := strings.TrimSpace(c.Query("order_no"))
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	events, pr, err := h.paymentAdmin.ListNotifications(c.Request.Context(), params, provider, orderNo, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentNotification, 0, len(events))
	for i := range events {
		e := events[i]
		out = append(out, dto.PaymentNotification{
			ID:              e.ID,
			Provider:        e.Provider,
			EventID:         e.EventID,
			OrderNo:         e.OrderNo,
			ProviderTradeNo: e.ProviderTradeNo,
			AmountCents:     e.AmountCents,
			Currency:        e.Currency,
			Verified:        e.Verified,
			Processed:       e.Processed,
			ProcessError:    e.ProcessError,
			RawBody:         e.RawBody,
			ReceivedAt:      e.ReceivedAt,
		})
	}

	if pr == nil {
		response.Paginated(c, out, int64(len(out)), page, pageSize)
		return
	}
	response.Paginated(c, out, pr.Total, pr.Page, pr.PageSize)
}

// ListProducts
// GET /api/v1/admin/payments/products?kind=&status=&search=&page=&page_size=
func (h *PaymentHandler) ListProducts(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	kind := strings.TrimSpace(c.Query("kind"))
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	products, pr, err := h.paymentAdmin.ListProducts(c.Request.Context(), params, kind, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentProduct, 0, len(products))
	for i := range products {
		p := products[i]
		out = append(out, dto.PaymentProduct{
			ID:                    p.ID,
			Kind:                  p.Kind,
			Name:                  p.Name,
			DescriptionMD:         p.DescriptionMD,
			Status:                p.Status,
			SortOrder:             p.SortOrder,
			Currency:              p.Currency,
			PriceCents:            p.PriceCents,
			GroupID:               p.GroupID,
			ValidityDays:          p.ValidityDays,
			CreditBalance:         p.CreditBalance,
			AllowCustomAmount:     p.AllowCustomAmount,
			MinAmountCents:        p.MinAmountCents,
			MaxAmountCents:        p.MaxAmountCents,
			SuggestedAmountsCents: p.SuggestedAmountsCents,
			ExchangeRate:          p.ExchangeRate,
			CreatedAt:             p.CreatedAt,
			UpdatedAt:             p.UpdatedAt,
		})
	}

	if pr == nil {
		response.Paginated(c, out, int64(len(out)), page, pageSize)
		return
	}
	response.Paginated(c, out, pr.Total, pr.Page, pr.PageSize)
}

// CreateProduct
// POST /api/v1/admin/payments/products
func (h *PaymentHandler) CreateProduct(c *gin.Context) {
	var req CreatePaymentProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	p, err := h.paymentAdmin.CreateProduct(c.Request.Context(), &service.CreatePaymentProductInput{
		Kind:                  req.Kind,
		Name:                  req.Name,
		DescriptionMD:         req.DescriptionMD,
		Status:                req.Status,
		SortOrder:             req.SortOrder,
		Currency:              req.Currency,
		PriceCents:            req.PriceCents,
		GroupID:               req.GroupID,
		ValidityDays:          req.ValidityDays,
		CreditBalance:         req.CreditBalance,
		AllowCustomAmount:     req.AllowCustomAmount,
		MinAmountCents:        req.MinAmountCents,
		MaxAmountCents:        req.MaxAmountCents,
		SuggestedAmountsCents: req.SuggestedAmountsCents,
		ExchangeRate:          req.ExchangeRate,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.PaymentProduct{
		ID:                    p.ID,
		Kind:                  p.Kind,
		Name:                  p.Name,
		DescriptionMD:         p.DescriptionMD,
		Status:                p.Status,
		SortOrder:             p.SortOrder,
		Currency:              p.Currency,
		PriceCents:            p.PriceCents,
		GroupID:               p.GroupID,
		ValidityDays:          p.ValidityDays,
		CreditBalance:         p.CreditBalance,
		AllowCustomAmount:     p.AllowCustomAmount,
		MinAmountCents:        p.MinAmountCents,
		MaxAmountCents:        p.MaxAmountCents,
		SuggestedAmountsCents: p.SuggestedAmountsCents,
		ExchangeRate:          p.ExchangeRate,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	})
}

// UpdateProduct
// PUT /api/v1/admin/payments/products/:id
func (h *PaymentHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	var req UpdatePaymentProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	p, err := h.paymentAdmin.UpdateProduct(c.Request.Context(), id, &service.UpdatePaymentProductInput{
		Kind:                  req.Kind,
		Name:                  req.Name,
		DescriptionMD:         req.DescriptionMD,
		Status:                req.Status,
		SortOrder:             req.SortOrder,
		Currency:              req.Currency,
		PriceCents:            req.PriceCents,
		GroupID:               req.GroupID,
		ValidityDays:          req.ValidityDays,
		CreditBalance:         req.CreditBalance,
		AllowCustomAmount:     req.AllowCustomAmount,
		MinAmountCents:        req.MinAmountCents,
		MaxAmountCents:        req.MaxAmountCents,
		SuggestedAmountsCents: req.SuggestedAmountsCents,
		ExchangeRate:          req.ExchangeRate,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentProduct{
		ID:                    p.ID,
		Kind:                  p.Kind,
		Name:                  p.Name,
		DescriptionMD:         p.DescriptionMD,
		Status:                p.Status,
		SortOrder:             p.SortOrder,
		Currency:              p.Currency,
		PriceCents:            p.PriceCents,
		GroupID:               p.GroupID,
		ValidityDays:          p.ValidityDays,
		CreditBalance:         p.CreditBalance,
		AllowCustomAmount:     p.AllowCustomAmount,
		MinAmountCents:        p.MinAmountCents,
		MaxAmountCents:        p.MaxAmountCents,
		SuggestedAmountsCents: p.SuggestedAmountsCents,
		ExchangeRate:          p.ExchangeRate,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	})
}

// DeleteProduct
// DELETE /api/v1/admin/payments/products/:id
func (h *PaymentHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	if err := h.paymentAdmin.DeleteProduct(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
