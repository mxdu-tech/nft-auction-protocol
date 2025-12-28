CREATE TABLE IF NOT EXISTS auctions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_id BIGINT UNSIGNED NOT NULL,
  contract_address VARCHAR(42) NOT NULL,
  auction_id BIGINT UNSIGNED NOT NULL,

  nft_contract VARCHAR(42) NOT NULL,
  token_id VARCHAR(80) NOT NULL,
  seller VARCHAR(42) NOT NULL,
  payment_token VARCHAR(42) NOT NULL,

  start_time DATETIME(3) NOT NULL,
  end_time DATETIME(3) NOT NULL,
  start_price DECIMAL(65,0) NOT NULL,

  status VARCHAR(20) NOT NULL,
  highest_bid DECIMAL(65,0) NOT NULL DEFAULT 0,
  highest_bidder VARCHAR(42) NOT NULL DEFAULT '',
  bid_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  winner VARCHAR(42) NOT NULL DEFAULT '',

  created_tx_hash VARCHAR(66) NOT NULL DEFAULT '',
  ended_tx_hash VARCHAR(66) NOT NULL DEFAULT '',
  created_block BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_block BIGINT UNSIGNED NOT NULL DEFAULT 0,

  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

  PRIMARY KEY (id),
  UNIQUE KEY uniq_chain_contract_auction (chain_id, contract_address, auction_id),
  KEY idx_chain_status_endtime (chain_id, status, end_time),
  KEY idx_seller (seller),
  KEY idx_nft (nft_contract, token_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS bids (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_id BIGINT UNSIGNED NOT NULL,
  auction_id BIGINT UNSIGNED NOT NULL,

  bidder VARCHAR(42) NOT NULL,
  amount DECIMAL(65,0) NOT NULL,
  payment_token VARCHAR(42) NOT NULL,

  tx_hash VARCHAR(66) NOT NULL,
  block_number BIGINT UNSIGNED NOT NULL,
  log_index INT UNSIGNED NOT NULL,
  bid_time DATETIME(3) NOT NULL,

  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

  PRIMARY KEY (id),
  UNIQUE KEY uniq_chain_tx_log (chain_id, tx_hash, log_index),
  KEY idx_auction_time (auction_id, bid_time),
  KEY idx_bidder (bidder)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS nft_assets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_id BIGINT UNSIGNED NOT NULL,
  nft_contract VARCHAR(42) NOT NULL,
  token_id VARCHAR(80) NOT NULL,

  owner VARCHAR(42) NOT NULL DEFAULT '',
  token_uri TEXT,
  name VARCHAR(255) NOT NULL DEFAULT '',
  image_url TEXT,
  collection_slug VARCHAR(255) NOT NULL DEFAULT '',

  floor_price DECIMAL(36,18) NOT NULL DEFAULT 0,
  floor_price_updated_at DATETIME(3) NULL,
  metadata_updated_at DATETIME(3) NULL,

  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

  PRIMARY KEY (id),
  UNIQUE KEY uniq_chain_nft (chain_id, nft_contract, token_id),
  KEY idx_owner (owner)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sync_cursors (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chain_id BIGINT UNSIGNED NOT NULL,
  contract_address VARCHAR(42) NOT NULL,
  event_name VARCHAR(64) NOT NULL,
  last_block BIGINT UNSIGNED NOT NULL DEFAULT 0,

  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

  PRIMARY KEY (id),
  UNIQUE KEY uniq_chain_contract_event (chain_id, contract_address, event_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;