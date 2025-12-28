package db

import (
	"context"
	"strings"
)

/* ========== Auction Report / Query ========== */

// 平台首页：进行中的拍卖
func (m *MySQL) ListOpenAuctions(
	ctx context.Context,
	chainID uint64,
	limit int,
	offset int,
) ([]Auction, error) {

	var auctions []Auction

	db := m.DB.WithContext(ctx).
		Where("chain_id = ? AND status = 'OPEN'", chainID).
		Order("end_time ASC")

	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.Find(&auctions).Error
	return auctions, err
}

// 已结束拍卖
func (m *MySQL) ListEndedAuctions(
	ctx context.Context,
	chainID uint64,
	limit int,
	offset int,
) ([]Auction, error) {

	var auctions []Auction

	db := m.DB.WithContext(ctx).
		Where("chain_id = ? AND status IN ('ENDED','CANCELLED')", chainID).
		Order("updated_at DESC")

	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}

	err := db.Find(&auctions).Error
	return auctions, err
}

// 拍卖详情页
func (m *MySQL) GetAuctionDetail(
	ctx context.Context,
	chainID uint64,
	contract string,
	auctionID uint64,
) (*Auction, error) {

	var a Auction

	err := m.DB.WithContext(ctx).
		Where(
			"chain_id = ? AND contract_address = ? AND auction_id = ?",
			chainID,
			strings.ToLower(contract),
			auctionID,
		).
		First(&a).Error

	if err != nil {
		return nil, err
	}

	return &a, nil
}

// 用户创建的拍卖
func (m *MySQL) ListAuctionsBySeller(
	ctx context.Context,
	chainID uint64,
	seller string,
) ([]Auction, error) {

	var auctions []Auction

	err := m.DB.WithContext(ctx).
		Where(
			"chain_id = ? AND seller = ?",
			chainID,
			strings.ToLower(seller),
		).
		Order("created_at DESC").
		Find(&auctions).Error

	return auctions, err
}

// 用户参与过的拍卖
func (m *MySQL) ListAuctionsBidByUser(
	ctx context.Context,
	chainID uint64,
	user string,
) ([]Auction, error) {

	var auctions []Auction

	err := m.DB.WithContext(ctx).
		Table("auctions a").
		Select("DISTINCT a.*").
		Joins(`
			JOIN bids b
			  ON a.chain_id = b.chain_id
			 AND a.auction_id = b.auction_id
		`).
		Where(
			"a.chain_id = ? AND b.bidder = ?",
			chainID,
			strings.ToLower(user),
		).
		Order("a.updated_at DESC").
		Scan(&auctions).Error

	return auctions, err
}
