package indexer

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/infra/db"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const CursorName = "auction_market_cursor"

type Indexer struct {
	cfg       *config.Config
	db        *db.MySQL
	client    *ethclient.Client
	contract  common.Address
	processor *Processor
}

func NewIndexer(cfg *config.Config, mysql *db.MySQL) (*Indexer, error) {
	cli, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(cfg.AuctionContract)

	return &Indexer{
		cfg:       cfg,
		db:        mysql,
		client:    cli,
		contract:  addr,
		processor: NewProcessor(mysql, cfg.ChainID, strings.ToLower(addr.Hex())),
	}, nil
}

func (i *Indexer) Start(ctx context.Context) {
	go i.loop(ctx)
}

func (i *Indexer) loop(ctx context.Context) {
	log.Printf(
		"[indexer] started chain_id=%d rpc=%s contract=%s",
		i.cfg.ChainID,
		i.cfg.RPCURL,
		i.contract.Hex(),
	)

	ticker := time.NewTicker(i.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[indexer] stopped")
			return
		case <-ticker.C:
			if err := i.tick(ctx); err != nil {
				log.Printf("[indexer] tick error: %v", err)
			}
		}
	}
}

func (i *Indexer) tick(ctx context.Context) error {
	chainID := i.cfg.ChainID
	contract := strings.ToLower(i.contract.Hex())

	// log.Printf(
	// 	"[indexer][debug] tick start chain_id=%d contract=%s cursor_name=%s init_depth=%d",
	// 	chainID,
	// 	contract,
	// 	CursorName,
	// 	i.cfg.InitScanDepth,
	// )

	last, err := i.db.GetCursor(chainID, contract, CursorName)

	// log.Printf(
	// 	"[indexer][debug] GetCursor result: last=%d err=%v",
	// 	last,
	// 	err,
	// )
	if err != nil {
		log.Printf("[indexer][debug] GetCursor error: %v", err)

		if err != db.ErrCursorNotFound {
			return err
		}

		latest, err := i.client.BlockNumber(ctx)
		if err != nil {
			return err
		}

		start := latest
		log.Print("[indexer][debug] latest block number: ", latest)
		if latest > i.cfg.InitScanDepth {
			start = latest - i.cfg.InitScanDepth
		}

		// log.Printf(
		// 	"[indexer][debug] cursor NOT FOUND → init cursor=%d (latest=%d, depth=%d)",
		// 	start,
		// 	latest,
		// 	i.cfg.InitScanDepth,
		// )

		return i.db.UpsertCursor(chainID, contract, CursorName, start)
	}

	// 2. 最新区块
	latest, err := i.client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	// confirmations
	if i.cfg.Confirmations > 0 && latest > i.cfg.Confirmations {
		latest -= i.cfg.Confirmations
	}

	if last >= latest {
		return nil
	}

	// 3. 控制 batch（防 RPC 限制）
	from := last + 1
	to := from + i.cfg.BatchSize - 1
	if to > latest {
		to = latest
	}

	log.Printf("[indexer] scan blocks [%d → %d] (latest=%d)", from, to, latest)

	// 4. 构造 filter
	ab := mustParseABI()
	topics := []common.Hash{
		ab.Events["AuctionCreated"].ID,
		ab.Events["BidPlaced"].ID,
		ab.Events["AuctionEnded"].ID,
		ab.Events["AuctionCancelled"].ID,
	}

	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{i.contract},
		Topics:    [][]common.Hash{topics},
	}

	logs, err := i.client.FilterLogs(ctx, q)
	if err != nil {
		return err
	}

	// 5. 处理 logs
	for _, lg := range logs {
		if err := i.processor.HandleLog(ctx, lg); err != nil {
			log.Printf(
				"[indexer] handle log error block=%d tx=%s err=%v",
				lg.BlockNumber,
				lg.TxHash.Hex(),
				err,
			)
		}
	}

	// 6. 推进 cursor
	if err := i.db.UpsertCursor(chainID, contract, CursorName, to); err != nil {
		return err
	}

	log.Printf("[indexer] cursor advanced → %d", to)

	return nil
}
