package middleware

import (
	"backend/internal/config"

	"github.com/gin-gonic/gin"
)

func InjectChainContext(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 把 chain_id 放进 gin context
		c.Set("chain_id", cfg.ChainID)

		c.Set("contract", cfg.AuctionContract)

		c.Next()
	}
}
