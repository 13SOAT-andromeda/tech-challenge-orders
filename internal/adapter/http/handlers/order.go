package handlers

import (
	"net/http"
	"strconv"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/middlewares"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/response"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
	"github.com/13SOAT-andromeda/tech-challenge-s1/pkg/converters"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service ports.OrderService
	usecase ports.OrderUseCase
}

func NewOrderHandler(service ports.OrderService, usecase ports.OrderUseCase) *OrderHandler {
	return &OrderHandler{service: service, usecase: usecase}
}

type CreateOrderRequest struct {
	VehicleKilometers int     `json:"vehicle_kilometers" binding:"required"`
	Note              *string `json:"note"`
	CustomerVehicleID uint    `json:"customer_vehicle_id" binding:"required"`
	CompanyID         uint    `json:"company_id" binding:"required"`
}

type CompleteAnalysisRequest struct {
	DiagnosticNote string             `json:"diagnostic_note"`
	Products       []domain.StockItem `json:"products" binding:"required"`
	Maintenances   []uint             `json:"maintenances" binding:"required"`
}

func (h *OrderHandler) Create(ctx *gin.Context) {
	var request CreateOrderRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userIdStr := middlewares.GetUserID(ctx)
	if userIdStr == "" {
		response.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, "Invalid User ID format")
		return
	}

	input := ports.CreateOrderInput{
		VehicleKilometers: request.VehicleKilometers,
		Note:              request.Note,
		CustomerVehicleID: request.CustomerVehicleID,
		CompanyID:         request.CompanyID,
	}

	created, err := h.usecase.CreateOrder(ctx.Request.Context(), uint(userId), input)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondCreated(ctx, created, "Order created successfully")
}

func (h *OrderHandler) Assign(ctx *gin.Context) {
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userIdStr := middlewares.GetUserID(ctx)
	if userIdStr == "" {
		response.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, "Invalid User ID format")
		return
	}

	if err := h.usecase.AssignOrder(ctx.Request.Context(), uint(orderID), uint(userId)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, "", "Order assigned successfully")
}

func (h *OrderHandler) CompleteAnalysis(ctx *gin.Context) {
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userIdStr := middlewares.GetUserID(ctx)
	if userIdStr == "" {
		response.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.RespondError(ctx, http.StatusUnauthorized, "Invalid User ID format")
		return
	}

	var request CompleteAnalysisRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	input := ports.CreateCompleteOrderAnalysisInput{
		DiagnosticNote: &request.DiagnosticNote,
		Products:       request.Products,
		Maintenances:   request.Maintenances,
	}

	if err := h.usecase.CompleteOrderAnalysis(ctx.Request.Context(), uint(orderID), uint(userId), input); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, "", "Order analysis completed successfully")
}

func (h *OrderHandler) GetAll(ctx *gin.Context) {
	u := ctx.Request.URL.Query()
	params := converters.ParamsToMap(u)

	orders, err := h.service.GetAll(ctx.Request.Context(), params)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(ctx, orders, "")
}

func (h *OrderHandler) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if order == nil {
		response.RespondError(ctx, http.StatusNotFound, "Order not found")
		return
	}
	response.RespondSuccess(ctx, order, "")
}

func (h *OrderHandler) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Order deleted successfully")
}

func (h *OrderHandler) ApproveOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.ApproveOrder(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Order approved successfully")
}

func (h *OrderHandler) RejectOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.RejectOrder(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Order rejected successfully")
}

func (h *OrderHandler) RequestApproval(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.RequestApproval(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Approval notification sent successfully")
}

func (h *OrderHandler) ArchiveOrder(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.ArchiveOrder(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Order archived successfully")
}

func (h *OrderHandler) StartWork(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.usecase.StartWorkOrder(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Work started successfully")
}

func (h *OrderHandler) GetInProgress(ctx *gin.Context) {
	u := ctx.Request.URL.Query()

	u.Add("sortdesc", "false")
	u.Add("orderby", "date_approved")
	u.Add("status", string(domain.OrderStatuses.IN_PROGRESS))

	params := converters.ParamsToMap(u)

	orders, err := h.service.GetAll(ctx.Request.Context(), params)
	if err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondSuccess(ctx, orders, "")
}

func (h *OrderHandler) CompleteWork(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.usecase.CompleteWorkOrder(ctx.Request.Context(), uint(id)); err != nil {
		response.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.RespondSuccess(ctx, id, "Work completed successfully")
}
