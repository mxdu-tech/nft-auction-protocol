package api

import (
	"backend/internal/infra/db"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, mysql *db.MySQL) {

	statsHandler := NewStatsHandler(mysql)
	auctionHandler := NewAuctionHandler(mysql)
	bidHandler := NewBidHandler(mysql)

	api := r.Group("/api")
	{
		api.GET("/stats", statsHandler.GetPlatformStats)

		api.GET("/auctions", auctionHandler.ListAuctions)
		api.GET("/auctions/:auctionId", auctionHandler.GetAuctionDetail)
		api.GET("/auctions/:auctionId/bids", bidHandler.ListBidsByAuction)
	}
}
