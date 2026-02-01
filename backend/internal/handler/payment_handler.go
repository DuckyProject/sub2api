package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// ListProducts
// GET /api/v1/payments/products?kind=subscription|balance
func (h *PaymentHandler) ListProducts(c *gin.Context) {
	kind := strings.TrimSpace(c.Query("kind"))
	products, err := h.paymentService.ListProducts(c.Request.Context(), kind)
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
	response.Success(c, out)
}

// CreateOrder
// POST /api/v1/payments/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req dto.CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	order, err := h.paymentService.CreateOrder(
		c.Request.Context(),
		subject.UserID,
		req.ProductID,
		strings.TrimSpace(req.Provider),
		req.ClientRequestID,
		req.Amount,
		req.AmountCents,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	resp := dto.CreatePaymentOrderResponse{
		Order: dto.PaymentOrder{
			ID:              order.ID,
			OrderNo:         order.OrderNo,
			UserID:          order.UserID,
			Kind:            order.Kind,
			ProductID:       order.ProductID,
			Status:          order.Status,
			Provider:        order.Provider,
			Currency:        order.Currency,
			AmountCents:     order.AmountCents,
			ClientRequestID: order.ClientRequestID,
			ProviderTradeNo: order.ProviderTradeNo,
			PayURL:          order.PayURL,
			ExpiresAt:       order.ExpiresAt,
			PaidAt:          order.PaidAt,
			FulfilledAt:     order.FulfilledAt,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
		},
	}
	response.Success(c, resp)
}

// ListOrders
// GET /api/v1/payments/orders?page=1&page_size=20&status=created|paid|fulfilled|cancelled|expired|failed
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	status := strings.TrimSpace(c.Query("status"))

	orders, pr, err := h.paymentService.ListOrders(c.Request.Context(), subject.UserID, params, status)
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

	response.PaginatedWithResult(c, out, &response.PaginationResult{
		Total:    pr.Total,
		Page:     pr.Page,
		PageSize: pr.PageSize,
		Pages:    pr.Pages,
	})
}

// GetOrder
// GET /api/v1/payments/orders/:order_no
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		response.BadRequest(c, "Missing order_no")
		return
	}

	order, err := h.paymentService.GetOrder(c.Request.Context(), subject.UserID, orderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentOrder{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		UserID:          order.UserID,
		Kind:            order.Kind,
		ProductID:       order.ProductID,
		Status:          order.Status,
		Provider:        order.Provider,
		Currency:        order.Currency,
		AmountCents:     order.AmountCents,
		ClientRequestID: order.ClientRequestID,
		ProviderTradeNo: order.ProviderTradeNo,
		PayURL:          order.PayURL,
		ExpiresAt:       order.ExpiresAt,
		PaidAt:          order.PaidAt,
		FulfilledAt:     order.FulfilledAt,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	})
}

// Notify
// POST /api/v1/payments/notify/:provider
func (h *PaymentHandler) Notify(c *gin.Context) {
	provider := strings.TrimSpace(c.Param("provider"))

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	statusCode, body, contentType, err := h.paymentService.HandleNotify(
		c.Request.Context(),
		provider,
		rawBody,
		c.Request.Header,
		c.Request.URL.Query(),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	c.Data(statusCode, contentType, []byte(body))
}
