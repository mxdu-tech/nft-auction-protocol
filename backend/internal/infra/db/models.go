package db

import "time"

type Auction struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	ChainID         uint64 `gorm:"not null;index:uniq_chain_contract_auction,unique,priority:1"`
	ContractAddress string `gorm:"size:42;not null;index:uniq_chain_contract_auction,unique,priority:2"`
	AuctionID       uint64 `gorm:"not null;index:uniq_chain_contract_auction,unique,priority:3"`

	NFTContract string `gorm:"size:42;not null;index:idx_nft,priority:1"`
	TokenID     string `gorm:"size:80;not null;index:idx_nft,priority:2"`
	Seller      string `gorm:"size:42;not null;index:idx_seller"`

	StartTime   time.Time `gorm:"not null"`
	EndTime     time.Time `gorm:"not null"`
	MinBidUsd18 string    `gorm:"size:100;not null"`

	Status           string `gorm:"size:20;not null;index:idx_chain_status_endtime,priority:2"`
	HighestBidToken  string `gorm:"size:42;not null"`
	HighestBidAmount string `gorm:"size:100;not null"`
	HighestBidUsd18  string `gorm:"size:100;not null"`
	HighestBidder    string `gorm:"size:42;not null"`
	BidCount         uint64 `gorm:"not null"`
	Winner           string `gorm:"size:42;not null"`

	CreatedTxHash string `gorm:"size:66;not null"`
	EndedTxHash   string `gorm:"size:66;not null"`
	CreatedBlock  uint64 `gorm:"not null"`
	UpdatedBlock  uint64 `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Bid struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	ChainID         uint64 `gorm:"not null;index:uniq_chain_tx_log,unique,priority:1"`
	ContractAddress string `gorm:"size:42;not null"`
	AuctionID       uint64 `gorm:"not null;index:idx_auction_time,priority:1"`

	Bidder    string `gorm:"size:42;not null;index:idx_bidder"`
	BidToken  string `gorm:"size:42;not null"`
	BidAmount string `gorm:"size:100;not null"`
	BidUsd18  string `gorm:"size:100;not null"`

	TxHash      string    `gorm:"size:66;not null;index:uniq_chain_tx_log,unique,priority:2"`
	BlockNumber uint64    `gorm:"not null"`
	LogIndex    uint      `gorm:"not null;index:uniq_chain_tx_log,unique,priority:3"`
	BidTime     time.Time `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type SyncCursor struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	ChainID         uint64 `gorm:"not null;index:uniq_chain_contract_cursor,unique,priority:1"`
	ContractAddress string `gorm:"size:42;not null;index:uniq_chain_contract_cursor,unique,priority:2"`
	EventName       string `gorm:"size:64;not null;index:uniq_chain_contract_cursor,unique,priority:3"`
	LastBlock       uint64 `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
