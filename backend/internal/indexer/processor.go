package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"backend/internal/infra/db"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type Processor struct {
	db       *db.MySQL
	gorm     *gorm.DB
	chainID  uint64
	contract string
	abi      abi.ABI
}

func NewProcessor(mysql *db.MySQL, chainID uint64, contractAddr string) *Processor {
	return &Processor{
		db:       mysql,
		gorm:     mysql.DB,
		chainID:  chainID,
		contract: strings.ToLower(contractAddr),
		abi:      mustParseABI(),
	}
}

func (p *Processor) HandleLog(ctx context.Context, lg types.Log) error {

	log.Printf(
		"HandleLog: tx=%s topic0=%s block=%d",
		lg.TxHash.Hex(),
		lg.Topics[0].Hex(),
		lg.BlockNumber,
	)

	if len(lg.Topics) == 0 {
		return nil
	}

	event, err := p.abi.EventByID(lg.Topics[0])
	if err != nil {
		return nil // 不是我们关心的事件
	}

	switch event.Name {
	case "AuctionCreated":
		return p.handleAuctionCreated(lg)
	case "BidPlaced":
		return p.handleBidPlaced(lg)
	case "AuctionEnded":
		return p.handleAuctionEnded(lg)
	case "AuctionCancelled":
		return p.handleAuctionCancelled(lg)
	default:
		return nil
	}
}

/* ================= AuctionCreated ================= */

func (p *Processor) handleAuctionCreated(lg types.Log) error {
	// 0. ctx（先用 Background，够用）
	ctx := context.Background()

	// 1. 解 data（非 indexed）
	var data struct {
		TokenId     *big.Int
		EndTime     *big.Int
		MinBidUsd18 *big.Int
	}

	if err := p.abi.UnpackIntoInterface(&data, "AuctionCreated", lg.Data); err != nil {
		return fmt.Errorf("unpack AuctionCreated data failed: %w", err)
	}

	// 2. topics：indexed 参数
	if len(lg.Topics) < 4 {
		return fmt.Errorf("invalid AuctionCreated topics len=%d", len(lg.Topics))
	}

	auctionId := new(big.Int).SetBytes(lg.Topics[1].Bytes()).Uint64()
	seller := common.BytesToAddress(lg.Topics[2].Bytes())
	nft := common.BytesToAddress(lg.Topics[3].Bytes())

	// 3. 防御性校验
	if data.TokenId == nil || data.EndTime == nil || data.MinBidUsd18 == nil {
		return fmt.Errorf("AuctionCreated data has nil field: %+v", data)
	}

	// 4. 构造 auction 对象
	auction := db.Auction{
		ChainID:         p.chainID,
		ContractAddress: p.contract,
		AuctionID:       auctionId,
		NFTContract:     nft.Hex(),
		TokenID:         data.TokenId.String(),
		Seller:          seller.Hex(),
		StartTime:       time.Unix(int64(lg.BlockNumber), 0), // 临时用 blockNumber
		EndTime:         time.Unix(data.EndTime.Int64(), 0),
		MinBidUsd18:     data.MinBidUsd18.String(),
		Status:          "OPEN",
		CreatedTxHash:   lg.TxHash.Hex(),
		CreatedBlock:    lg.BlockNumber,
	}

	// 5. 写库
	// return p.db.InsertAuctionDebug(ctx, &auction)
	return p.db.UpsertAuctionWithLog(ctx, &auction)
}

/* ================= BidPlaced ================= */

func (p *Processor) handleBidPlaced(lg types.Log) error {
	// 1. 解非 indexed data
	var data struct {
		BidToken  common.Address
		BidAmount *big.Int
		BidUsd18  *big.Int
	}

	if err := p.abi.UnpackIntoInterface(&data, "BidPlaced", lg.Data); err != nil {
		return fmt.Errorf("unpack BidPlaced failed: %w", err)
	}

	// 2. 解 indexed topics
	if len(lg.Topics) < 3 {
		return fmt.Errorf("invalid BidPlaced topics len=%d", len(lg.Topics))
	}

	auctionID := new(big.Int).SetBytes(lg.Topics[1].Bytes()).Uint64()
	bidder := common.BytesToAddress(lg.Topics[2].Bytes())

	// 3. 防御性校验
	if data.BidAmount == nil || data.BidUsd18 == nil {
		return fmt.Errorf("BidPlaced data has nil field: %+v", data)
	}

	// 4. 插入 bid
	bid := &db.Bid{
		ChainID:         p.chainID,
		ContractAddress: p.contract,
		AuctionID:       auctionID,
		Bidder:          strings.ToLower(bidder.Hex()),
		BidToken:        strings.ToLower(data.BidToken.Hex()),
		BidAmount:       data.BidAmount.String(),
		BidUsd18:        data.BidUsd18.String(),
		TxHash:          lg.TxHash.Hex(),
		BlockNumber:     lg.BlockNumber,
		BidTime:         time.Now().UTC(),
	}

	if err := p.db.InsertBid(context.Background(), bid); err != nil {
		return fmt.Errorf("insert bid failed: %w", err)
	}

	// 5. 调 Repo
	return p.db.UpdateAuctionOnBid(
		context.Background(),
		p.chainID,
		p.contract,
		auctionID,
		strings.ToLower(data.BidToken.Hex()), // address → string
		data.BidAmount.String(),
		data.BidUsd18.String(),
		strings.ToLower(bidder.Hex()),
		lg.BlockNumber,
	)
}

//* ================= AuctionEnded ================= */

func (p *Processor) handleAuctionEnded(lg types.Log) error {
	ctx := context.Background()

	// 1. 解 data（全部是 non-indexed）
	var ev struct {
		Winner    common.Address
		PayToken  common.Address
		PayAmount *big.Int
		PayUsd18  *big.Int
	}

	if err := p.abi.UnpackIntoInterface(&ev, "AuctionEnded", lg.Data); err != nil {
		return fmt.Errorf("unpack AuctionEnded data failed: %w", err)
	}

	// 2. 解 indexed：auctionId
	if len(lg.Topics) < 2 {
		return fmt.Errorf("invalid AuctionEnded topics len=%d", len(lg.Topics))
	}

	auctionID := new(big.Int).SetBytes(lg.Topics[1].Bytes()).Uint64()

	// 3. 调 repo（不在 processor 里写 SQL）
	if err := p.db.MarkAuctionEnded(
		ctx,
		p.chainID,
		p.contract,
		auctionID,
		strings.ToLower(ev.Winner.Hex()),
		lg.TxHash.Hex(),
		lg.BlockNumber,
	); err != nil {
		return err
	}

	return nil
}

/* ================= AuctionCancelled ================= */

func (p *Processor) handleAuctionCancelled(lg types.Log) error {
	// 1. 校验 topics
	if len(lg.Topics) < 2 {
		return fmt.Errorf("AuctionCancelled: invalid topics len=%d", len(lg.Topics))
	}

	// 2. 解析 auctionId（唯一字段）
	auctionID := new(big.Int).SetBytes(lg.Topics[1].Bytes()).Uint64()

	// 3. 调用 repo（不在这里写 SQL）
	return p.db.MarkAuctionCancelled(
		context.Background(),
		p.chainID,
		p.contract,
		auctionID,
		lg.TxHash.Hex(),
		lg.BlockNumber,
	)
}
