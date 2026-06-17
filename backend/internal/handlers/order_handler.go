package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus-stack/backend/internal/models"
)

type OrderHandler struct {
	repo *models.OrderRepository
}

func NewOrderHandler(repo *models.OrderRepository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) List(c *gin.Context) {
	status := c.Query("status")
	orders, err := h.repo.List(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	c.JSON(http.StatusOK, gin.H{"data": orders, "count": len(orders)})
}

func (h *OrderHandler) Get(c *gin.Context) {
	id := c.Param("id")
	o, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}
	if o == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": o})
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": o})
}
