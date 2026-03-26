# NFT Auction Protocol (Upgradeable + Off-chain Indexer)

A full-stack Web3 NFT auction system built on Ethereum (Sepolia), featuring an upgradeable smart contract architecture and a production-style off-chain indexing backend.

This project demonstrates an end-to-end pipeline:

On-chain contract → Event indexing → Structured storage → Queryable API

---

## Architecture Overview

The system consists of two major components:

### 1. On-chain (Smart Contracts)
- NFT Auction Market (UUPS Upgradeable)
- Supports ETH and ERC20 (USDC) bidding
- Chainlink Price Feeds for fair USD-based comparison

### 2. Off-chain Backend (Indexer + API)
- Real-time blockchain event indexing
- Persistent storage in MySQL
- REST API for frontend consumption

---

## Key Features

### Smart Contract Layer

- UUPS upgradeable architecture (ERC1967 standard)
- Dual-token bidding (ETH + USDC)
- Chainlink Oracle integration (ETH/USD, USDC/USD)
- Unified USD-based bid comparison
- Full auction lifecycle:
  - Create auction
  - Place bid
  - End auction
  - Settlement

### Backend Layer

- Event-driven indexer (based on go-ethereum)
- Automatic chain synchronization with cursor tracking
- Resumable indexing (fault-tolerant)
- REST API abstraction for frontend
- Clean separation of domain logic and infrastructure

---

## Technical Highlights

### Upgradeable Smart Contract Design
- Implemented UUPS proxy pattern
- Ensured storage layout compatibility across versions
- Verified upgrade safety via V1 → V2 migration

### Multi-token Price Normalization
- Integrated Chainlink Price Feeds
- Converted all bids into USD (18 decimals)
- Eliminated unfair advantages across token types

### Event-driven Backend Architecture
- Subscribed to contract events:
  - AuctionCreated
  - BidPlaced
  - AuctionEnded
  - AuctionCancelled
- Parsed logs into structured domain models
- Stored into relational database

### Fault-tolerant Indexing
- Maintained sync_cursors table
- Enabled restart-safe indexing
- Supports backfilling historical blocks

### Clean API Abstraction
Frontend interacts with backend without touching blockchain:

- /api/auctions
- /api/auctions/:id
- /api/auctions/:id/bids
- /api/stats

---

## Project Structure

nft-auction-dapp/
├── onchain/        # Smart contracts (Hardhat)
├── backend/        # Indexer + API (Go)

### On-chain
contracts/
  ├── MyNFT.sol
  ├── NFTAuctionMarketUUPS_V1.sol
  ├── NFTAuctionMarketUUPS_V2.sol

### Backend
cmd/server         # Entry point
internal/indexer   # Chain event processing
internal/api       # HTTP handlers
internal/infra/db  # DB models & repository layer
internal/config    # Configuration

---

## Tech Stack

### Smart Contracts
- Solidity
- Hardhat
- OpenZeppelin (UUPSUpgradeable)
- Chainlink Oracle

### Backend
- Go
- Gin
- GORM
- MySQL
- go-ethereum

### Infra
- Docker Compose

---

## Deployment (Sepolia)

Proxy (Auction): 0x085Ab5880dff3EDaE319948946E1df5bF683934e  
V2 Implementation: 0x6E5F7Ba248F6deaC57ba0cfB565C0c7f24240f94  
MyNFT: 0x7613Fea95c816346beaBE7132A29BC4Ab14BE83b  

---

## Testing

### Local (Hardhat)

Covers:
- Auction creation
- ETH vs USDC bidding logic
- Price comparison via oracle
- Auction settlement
- Upgradeability validation

Run:
npx hardhat test

---

### Testnet (Sepolia)

Validated full workflow:
- Deploy NFT contract
- Mint NFT
- Approve marketplace
- Create auction
- Multi-round bidding
- Oracle price verification
- Auction settlement
- Contract upgrade (V1 → V2)

---

## Running Backend

docker-compose up

or

go run cmd/server/main.go

---

## API Endpoints

GET /health  
GET /api/stats  
GET /api/auctions  
GET /api/auctions/:id  
GET /api/auctions/:id/bids  

---

## Screenshots

See /screenshots folder.

---

## What This Project Demonstrates

- Production-style smart contract architecture
- Upgradeable protocol design
- Cross-token economic logic (USD normalization)
- Event-driven backend system
- End-to-end Web3 system integration

---

## Future Improvements

- WebSocket real-time updates
- Frontend integration (React / Next.js)
- Multi-chain support
- Advanced auction types (Dutch, sealed-bid)

---

## Author

M.X Du

Smart Contract Engineer (EVM)  
Focused on protocol design, upgradeability, and backend indexing systems.
