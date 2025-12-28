package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ChainID         uint64
	RPCURL          string
	AuctionContract string

	PollInterval  time.Duration
	Confirmations uint64

	InitScanDepth uint64
	BatchSize     uint64
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Env:  getenv("APP_ENV", "dev"),
		Port: getenv("APP_PORT", "8080"),

		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "3306"),
		DBUser:     getenv("DB_USER", "nft"),
		DBPassword: getenv("DB_PASSWORD", "nftpassword"),
		DBName:     getenv("DB_NAME", "nft_auction"),

		RPCURL:          getenv("RPC_URL", "http://127.0.0.1:8545"),
		AuctionContract: getenv("AUCTION_CONTRACT", ""),
		Confirmations:   uint64(getenvInt("CONFIRMATIONS", 0)),

		InitScanDepth: uint64(getenvInt("INIT_SCAN_DEPTH", 0)),
		BatchSize:     uint64(getenvInt("BATCH_SIZE", 0)),
	}

	cfg.ChainID = uint64(getenvInt("CHAIN_ID", 31337))

	sec := getenvInt("POLL_INTERVAL_SECONDS", 3)
	cfg.PollInterval = time.Duration(sec) * time.Second

	return cfg
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
