package handlers

import (
	"net/http"
	"strconv"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/middlewares"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/http/response"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	orderport "github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	usecase ports.OrderUseCase
}

func NewOrderHandler(usecase ports.OrderUseCase) *OrderHandler {
	return &OrderHandler{usecase: usecase}
}

type CreateOrderRequest struct {
	VehicleKilometers int     `json:"vehicle_kilometers" binding:"required"`
	Note              *string `json:"note"`
	CustomerVehicleID int64   `json:"customer_vehicle_id" binding:"required"`
	EmployeeID        int64   `json:"employee_id" binding:"required"`
	CompanyID         int64   `json:"company_id" binding:"required"`
}

type StockItemRequest struct {
	ID       string `json:"id" binding:"required"`
	Quantity uint   `json:"quantity" binding:"required"`
}

type CompleteAnalysisRequest struct {
	DiagnosticNote string             `json:"diagnostic_note"`
	Products       []StockItemRequest `json:"products" binding:"required"`
	Maintenances   []int64            `json:"maintenances" binding:"required"`
}

func (h *OrderHandler) Create(ctx *gin.Context) {
	var request CreateOrderRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := middlewares.GetUserIDInt64(ctx)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	input := ports.CreateOrderInput{
		VehicleKilometers: request.VehicleKilometers,
		Note:              request.Note,
		CustomerVehicleID: request.CustomerVehicleID,
		EmployeeID:        request.EmployeeID,
		CompanyID:         request.CompanyID,
	}

	created, err := h.usecase.CreateOrder(ctx.Request.Context(), userID, input)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondCreated(ctx, created, "Order created successfully")
}

func (h *OrderHandler) Assign(ctx *gin.Context) {
	orderID := ctx.Param("id")
	userID, err := middlewares.GetUserIDInt64(ctx)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.usecase.AssignOrder(ctx.Request.Context(), orderID, userID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, "", "Order assigned successfully")
}

func (h *OrderHandler) CompleteAnalysis(ctx *gin.Context) {
	orderID := ctx.Param("id")

	userID, err := middlewares.GetUserIDInt64(ctx)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var request CompleteAnalysisRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	products := make([]domain.StockItem, 0, len(request.Products))
	for _, p := range request.Products {
		products = append(products, domain.StockItem{ID: p.ID, Quantity: p.Quantity})
	}

	input := ports.CreateCompleteOrderAnalysisInput{
		DiagnosticNote: &request.DiagnosticNote,
		Products:       products,
		Maintenances:   request.Maintenances,
	}

	if err := h.usecase.CompleteOrderAnalysis(ctx.Request.Context(), orderID, userID, input); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, "", "Order analysis completed successfully")
}

func (h *OrderHandler) GetAll(ctx *gin.Context) {
	page := parsePage(ctx)
	result, err := h.usecase.GetAllOrders(ctx.Request.Context(), page)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(ctx, result, "")
}

func (h *OrderHandler) GetByID(ctx *gin.Context) {
	orderID := ctx.Param("id")
	detail, err := h.usecase.GetOrderByID(ctx.Request.Context(), orderID)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			response.RespondError(ctx, http.StatusNotFound, err.Error())
			return
		}
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(ctx, detail, "")
}

func (h *OrderHandler) GetInProgress(ctx *gin.Context) {
	page := parsePage(ctx)
	result, err := h.usecase.GetInProgressOrders(ctx.Request.Context(), page)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(ctx, result, "")
}

func parsePage(ctx *gin.Context) orderport.Page {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	return orderport.Page{
		Limit:  int32(limit),
		Cursor: ctx.Query("cursor"),
	}
}

func (h *OrderHandler) ApproveOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if err := h.usecase.ApproveOrder(ctx.Request.Context(), orderID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, orderID, "Order approved successfully")
}

func (h *OrderHandler) RejectOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if err := h.usecase.RejectOrder(ctx.Request.Context(), orderID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, orderID, "Order rejected successfully")
}

func (h *OrderHandler) RequestApproval(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if err := h.usecase.RequestApproval(ctx.Request.Context(), orderID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, orderID, "Approval notification sent successfully")
}

func (h *OrderHandler) ArchiveOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if err := h.usecase.ArchiveOrder(ctx.Request.Context(), orderID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, orderID, "Order archived successfully")
}

func (h *OrderHandler) CompleteWork(ctx *gin.Context) {
	orderID := ctx.Param("id")

	if err := h.usecase.CompleteWorkOrder(ctx.Request.Context(), orderID); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, orderID, "Work completed successfully")
}
