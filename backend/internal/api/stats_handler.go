package api

import (
	"net/http"

	"backend/internal/infra/db"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	db *db.MySQL
}

func NewStatsHandler(db *db.MySQL) *StatsHandler {
	return &StatsHandler{db: db}
}

func (h *StatsHandler) GetPlatformStats(c *gin.Context) {
	chainID := c.MustGet("chain_id").(uint64)

	stats, err := h.db.GetPlatformStats(c.Request.Context(), chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
