package db

import (
	"context"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/* ---------------- Auction ---------------- */
func (m *MySQL) InsertAuctionDebug(ctx context.Context, a *Auction) error {
	tx := m.DB.WithContext(ctx).Create(a)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		log.Printf("[WARN] auction NOT inserted: chain=%d contract=%s auctionId=%d",
			a.ChainID, a.ContractAddress, a.AuctionID)
	} else {
		log.Printf("[OK] auction inserted: auctionId=%d", a.AuctionID)
	}

	return nil
}

// AuctionCreated：插入 auction（已存在则忽略）
func (m *MySQL) UpsertAuctionWithLog(ctx context.Context, a *Auction) error {
	tx := m.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "chain_id"},
				{Name: "contract_address"},
				{Name: "auction_id"},
			},
			DoNothing: true,
		}).
		Create(a)

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		log.Printf(
			"[SKIP] auction already exists: chain=%d contract=%s auctionId=%d",
			a.ChainID,
			a.ContractAddress,
			a.AuctionID,
		)
	} else {
		log.Printf(
			"[OK] auction inserted: auctionId=%d tokenId=%s",
			a.AuctionID,
			a.TokenID,
		)
	}

	return nil
}

// BidPlaced：更新最高出价信息（带日志）
func (m *MySQL) UpdateAuctionOnBid(
	ctx context.Context,
	chainID uint64,
	contract string,
	auctionID uint64,
	bidToken string,
	bidAmount string,
	bidUsd18 string,
	bidder string,
	blockNumber uint64,
) error {

	tx := m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where(
			"chain_id = ? AND contract_address = ? AND auction_id = ?",
			chainID,
			contract,
			auctionID,
		).
		Updates(map[string]interface{}{
			"highest_bid_token":  bidToken,
			"highest_bid_amount": bidAmount,
			"highest_bid_usd18":  bidUsd18,
			"highest_bidder":     bidder,
			"bid_count":          gorm.Expr("bid_count + 1"),
			"updated_block":      blockNumber,
		})

	// 1. SQL 错误
	if tx.Error != nil {
		log.Printf(
			"[ERR] UpdateAuctionOnBid failed: chain=%d contract=%s auctionId=%d err=%v",
			chainID, contract, auctionID, tx.Error,
		)
		return tx.Error
	}

	// 2. 没有命中任何 auction
	if tx.RowsAffected == 0 {
		log.Printf(
			"[WARN] UpdateAuctionOnBid skipped (auction not found): chain=%d contract=%s auctionId=%d",
			chainID, contract, auctionID,
		)
		return nil
	}

	// 3. 正常更新
	log.Printf(
		"[OK] auction updated on bid: auctionId=%d bidder=%s amount=%s token=%s",
		auctionID, bidder, bidAmount, bidToken,
	)

	return nil
}

func (m *MySQL) MarkAuctionEnded(
	ctx context.Context,
	chainID uint64,
	contract string,
	auctionID uint64,
	winner string,
	endTx string,
	blockNumber uint64,
) error {

	tx := m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where(
			"chain_id = ? AND contract_address = ? AND auction_id = ?",
			chainID,
			contract,
			auctionID,
		).
		Updates(map[string]interface{}{
			"status":        "ENDED",
			"winner":        winner,
			"ended_tx_hash": endTx,
			"updated_block": blockNumber,
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		log.Printf(
			"[WARN] AuctionEnded NOT updated: chain=%d contract=%s auctionId=%d",
			chainID, contract, auctionID,
		)
	} else {
		log.Printf(
			"[OK] AuctionEnded updated: auctionId=%d winner=%s",
			auctionID, winner,
		)
	}

	return nil
}

// AuctionCancelled
func (m *MySQL) MarkAuctionCancelled(
	ctx context.Context,
	chainID uint64,
	contract string,
	auctionID uint64,
	cancelTx string,
	blockNumber uint64,
) error {

	tx := m.DB.WithContext(ctx).
		Model(&Auction{}).
		Where(
			"chain_id = ? AND contract_address = ? AND auction_id = ?",
			chainID,
			contract,
			auctionID,
		).
		Updates(map[string]interface{}{
			"status":        "CANCELLED",
			"ended_tx_hash": cancelTx,
			"updated_block": blockNumber,
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		log.Printf(
			"[WARN] auction cancel NOT updated: chain=%d contract=%s auctionId=%d",
			chainID, contract, auctionID,
		)
	} else {
		log.Printf(
			"[OK] auction cancelled: auctionId=%d",
			auctionID,
		)
	}

	return nil
}

/* ---------------- Bid ---------------- */

// BidPlaced：插入 bid（带日志）
func (m *MySQL) InsertBid(ctx context.Context, b *Bid) error {
	tx := m.DB.WithContext(ctx).Create(b)

	// 1. SQL 错误
	if tx.Error != nil {
		log.Printf(
			"[ERR] bid insert failed: chain=%d contract=%s auctionId=%d bidder=%s err=%v",
			b.ChainID,
			b.ContractAddress,
			b.AuctionID,
			b.Bidder,
			tx.Error,
		)
		return tx.Error
	}

	// 2. 理论上不会出现，但防御性处理
	if tx.RowsAffected == 0 {
		log.Printf(
			"[WARN] bid NOT inserted (rows=0): chain=%d contract=%s auctionId=%d bidder=%s",
			b.ChainID,
			b.ContractAddress,
			b.AuctionID,
			b.Bidder,
		)
		return nil
	}

	// 3. 正常插入
	log.Printf(
		"[OK] bid inserted: auctionId=%d bidder=%s amount=%s token=%s tx=%s",
		b.AuctionID,
		b.Bidder,
		b.BidAmount,
		b.BidToken,
		b.TxHash,
	)

	return nil
}
