package api

import (
	"net/http"
	"strconv"

	"backend/internal/infra/db"

	"github.com/gin-gonic/gin"
)

type AuctionHandler struct {
	db *db.MySQL
}

func NewAuctionHandler(db *db.MySQL) *AuctionHandler {
	return &AuctionHandler{db: db}
}

// GET /api/auctions?status=open|ended
func (h *AuctionHandler) ListAuctions(c *gin.Context) {
	chainID := c.MustGet("chain_id").(uint64)
	status := c.DefaultQuery("status", "open")

	var (
		auctions []db.Auction
		err      error
	)

	switch status {
	case "open":
		auctions, err = h.db.ListOpenAuctions(c.Request.Context(), chainID, 50, 0)
	case "ended":
		auctions, err = h.db.ListEndedAuctions(c.Request.Context(), chainID, 50, 0)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, auctions)
}

// GET /api/auctions/:id
func (h *AuctionHandler) GetAuctionDetail(c *gin.Context) {
	chainID := c.MustGet("chain_id").(uint64)
	auctionID, _ := strconv.ParseUint(c.Param("auctionId"), 10, 64)
	contract := c.MustGet("contract").(string)

	auction, err := h.db.GetAuctionDetail(
		c.Request.Context(),
		chainID,
		contract,
		auctionID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auction not found"})
		return
	}

	c.JSON(http.StatusOK, auction)
}
