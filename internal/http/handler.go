package http

import (
	"net/http"
	"strconv"

	"sistem-manajemen-armada/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.LocationService
}

func NewHandler(svc *service.LocationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v := r.Group("/vehicles")
	{
		v.GET("/:vehicle_id/location", h.GetLatestLocation)
		v.GET("/:vehicle_id/history", h.GetHistory)
	}
}

func (h *Handler) GetLatestLocation(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	loc, err := h.svc.GetLatest(c.Request.Context(), vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	c.JSON(http.StatusOK, loc)
}

func (h *Handler) GetHistory(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err1 := strconv.ParseInt(startStr, 10, 64)
	end, err2 := strconv.ParseInt(endStr, 10, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start/end query param"})
		return
	}

	locations, err := h.svc.GetHistory(c.Request.Context(), vehicleID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query history"})
		return
	}

	c.JSON(http.StatusOK, locations)
}
