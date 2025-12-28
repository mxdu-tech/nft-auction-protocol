package db

import "context"

type PlatformStats struct {
	AuctionTotal int64 `json:"auction_total"`
	AuctionOpen  int64 `json:"auction_open"`
	AuctionEnded int64 `json:"auction_ended"`
	BidTotal     int64 `json:"bid_total"`
}

func (m *MySQL) GetPlatformStats(
	ctx context.Context,
	chainID uint64,
) (*PlatformStats, error) {

	var s PlatformStats

	if err := m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where("chain_id = ?", chainID).
		Count(&s.AuctionTotal).Error; err != nil {
		return nil, err
	}

	m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where("chain_id = ? AND status = 'OPEN'", chainID).
		Count(&s.AuctionOpen)

	m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where("chain_id = ? AND status IN ('ENDED','CANCELLED')", chainID).
		Count(&s.AuctionEnded)

	m.DB.WithContext(ctx).
		Model(&Bid{}).
		Where("chain_id = ?", chainID).
		Count(&s.BidTotal)

	return &s, nil
}
