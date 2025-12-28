package db

import (
	"context"
)

func (m *MySQL) ListBidsByAuction(
	ctx context.Context,
	chainID uint64,
	auctionID uint64,
) ([]Bid, error) {

	var bids []Bid

	err := m.DB.WithContext(ctx).
		Where(
			"chain_id = ? AND auction_id = ?",
			chainID,
			auctionID,
		).
		Order("block_number ASC, log_index ASC").
		Find(&bids).Error

	return bids, err
}
