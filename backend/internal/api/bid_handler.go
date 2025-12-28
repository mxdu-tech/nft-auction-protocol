package api

import (
	"net/http"
	"strconv"

	"backend/internal/infra/db"

	"github.com/gin-gonic/gin"
)

type BidHandler struct {
	db *db.MySQL
}

func NewBidHandler(db *db.MySQL) *BidHandler {
	return &BidHandler{db: db}
}

// GET /api/auctions/:id/bids
func (h *BidHandler) ListBidsByAuction(c *gin.Context) {
	chainID := c.MustGet("chain_id").(uint64)
	auctionID, _ := strconv.ParseUint(c.Param("auctionId"), 10, 64)

	bids, err := h.db.ListBidsByAuction(
		c.Request.Context(),
		chainID,
		auctionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bids)
}
